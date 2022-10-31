import json
import socket
import asyncio
import time
import traceback
import platform
import ssl

import pynvml   # pip install nvidia-ml-py
import websockets.client   # pip install websockets, https://github.com/aaugustin/websockets/issues/974#issuecomment-932727653
import psutil

SERVER_ADDR = "pris.ssdk.icu"
SERVER_PORT = 443
SEND_INTERVAL = 6       # 默认发送间隔，单位为秒
RETRY_INTERVAL = 10     # 建立连接失败时重试的等待间隔，单位为秒
UNIT = 1024 ** 3        # 使用 GB 作为单位


class Client:
    def __init__(self):
        self.local_ip = Client.extract_ip()
        pynvml.nvmlInit()   # 初始化
        self.gpu_count = pynvml.nvmlDeviceGetCount()    # 获取 Nvidia GPU 块数
        handle = pynvml.nvmlDeviceGetHandleByIndex(0)
        self.gpu_model = str(pynvml.nvmlDeviceGetName(handle))
        self.platform = platform.system()
        # self.get_gpu_process_json()

    def __del__(self):
        pynvml.nvmlShutdown()   # 关闭管理工具

    @staticmethod
    def extract_ip():
        st = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        try:       
            st.connect(('10.255.255.255', 1))
            ip = st.getsockname()[0]
            print(ip)
        except Exception:
            ip = '127.0.0.1'
        finally:
            st.close()
        return ip

    def get_gpu_process_json(self):
        t_gpu_info=[]
        for i in range(self.gpu_count):
            handle = pynvml.nvmlDeviceGetHandleByIndex(i)
            # if(self.platform =="Windows"):
            #     pidAllInfo = pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle)
            # else:
            #     pidAllInfo = pynvml.nvmlDeviceGetComputeRunningProcesses(handle)
            pidAllInfo = pynvml.nvmlDeviceGetComputeRunningProcesses(handle)
            if(len(pidAllInfo)==0):
                pidAllInfo = pynvml.nvmlDeviceGetGraphicsRunningProcesses(handle)

            for pidInfo in pidAllInfo:
                pidUser = psutil.Process(pidInfo.pid).username()
                pidName = psutil.Process(pidInfo.pid).cmdline()
                if isinstance(pidName,list):
                    pidName=' '.join(pidName)
                gpu_used=pidInfo.usedGpuMemory
                if gpu_used is not None:
                    gpu_used/=UNIT
                else:
                    gpu_used=0
                t_gpu_info.append({"Id": i, "Ip": self.local_ip, "User": pidUser, "MemUsed": gpu_used, "Name": pidName})
        if(len(t_gpu_info)==0):
            t_gpu_info=[{}]
        return json.dumps(
            {
                "Type": 2,
                "Data": t_gpu_info
            }
        )

    def get_gpu_stat_json(self):
        gpu_stats = []

        # gpuDriveInfo = pynvml.nvmlSystemGetDriverVersion()
        # print("Driver 版本: ", str(gpuDriveInfo)) # 驱动版本信息

        for i in range(self.gpu_count):
            handle = pynvml.nvmlDeviceGetHandleByIndex(i)   # 获取GPU i的handle，后续通过handle来处理

            memoryInfo = pynvml.nvmlDeviceGetMemoryInfo(handle) # 通过handle获取GPU i的信息
            gpuName = str(pynvml.nvmlDeviceGetName(handle))
            gpuTemperature = pynvml.nvmlDeviceGetTemperature(handle, 0)
            gpuFanSpeed = pynvml.nvmlDeviceGetFanSpeed(handle)
            gpuPowerState = pynvml.nvmlDeviceGetPowerState(handle)
            gpuUtilRate = pynvml.nvmlDeviceGetUtilizationRates(handle).gpu
            gpuMemoryRate = pynvml.nvmlDeviceGetUtilizationRates(handle).memory

            gpu_stats.append(
                {
                    "Ip": self.local_ip,
                    "GpuId": i,
                    # "GpuName": gpuName,
                    "MemTotal": memoryInfo.total / UNIT,
                    "MemUsed": memoryInfo.used / UNIT,
                    "MemFree": memoryInfo.free / UNIT,
                    "GpuTemp": gpuTemperature,
                    "GpuFanSpeed": gpuFanSpeed,
                    "GpuPowerStat": gpuPowerState, # 供电水平
                    "GpuUtilRate": gpuUtilRate,     # 计算核心满速使用率：
                    "GpuMemRate": gpuMemoryRate,  # 内存读写满速使用率
                    "Time": int(round(time.time() * 1000))      # Unix 时间戳 (13位)
                }
            )

        return json.dumps(
            {
                "Type": 1,
                "Data": gpu_stats
            }
        )

    async def handshake(self, websocket):
        await websocket.send(json.dumps(
                {
                    "Type": 0,
                    "Data": [{"Ip": self.local_ip,
                              "Name": socket.gethostname(),
                              "Cnt": self.gpu_count,
                              "Model": self.gpu_model}]
                }
            ))
        # response_str = await websocket.recv()
        # print(response_str)
        return True

    async def send_data(self, websocket):
        while True:
            try:
                need_update = await websocket.recv()
                need_update=json.loads(need_update)
                print(need_update)
                Type=need_update['Type']
                if Type==-1:
                    await websocket.send(self.get_gpu_stat_json())
                if Type==2:
                    await websocket.send(self.get_gpu_process_json())
                # time.sleep(SEND_INTERVAL)
            except:
                print("Connection is Closed")
                data = None
                break

    async def launch(self):
        while True:
            try:
                #todo kluge
                #HIGHLY INSECURE
                ssl_context = ssl.create_default_context()
                ssl_context.check_hostname = False
                ssl_context.verify_mode = ssl.CERT_NONE
                async with websockets.client.connect(f"wss://{SERVER_ADDR}:{SERVER_PORT}/node", ssl=ssl_context) as websocket:
                    print("Connection succeeded!")
                    await self.handshake(websocket)
                    await self.send_data(websocket)
            except Exception:
                print(f"[ERROR]\n{traceback.print_exc()}, will retry in {RETRY_INTERVAL} seconds.")
                time.sleep(RETRY_INTERVAL)


if __name__ == "__main__":
    c = Client()
    # print(c.get_gpu_process_json())
    try:
        asyncio.get_event_loop().run_until_complete(c.launch())
    finally:
        del c
