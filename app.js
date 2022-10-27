// app.js
var socket = {};
App({

  watch_gpuInfo: function (method) {
    //监听函数
    var obj = this.globalData
    Object.defineProperty(obj, 'gpuInfo', {
      configurable: true,
      enumerable: true,
      set: function (value) {
        this._gpuInfo = value;
        method(value);
      },

      get: function () {
        if (this._gpuInfo == undefined) {
          return {}
        } else {
          return this._gpuInfo
        }
      }
    })
  },
  watch_machineIPs: function (method) {
    //监听函数
    var obj = this.globalData
    Object.defineProperty(obj, 'machineIPs', {
      configurable: true,
      enumerable: true,
      set: function (value) {
        this._machineIPs = value;
        method(value);
      },

      get: function () {
        if (this._machineIPs == undefined) {
          return []
        } else {
          return this._machineIPs
        }
      }
    })
  },
  watch_gpuProc: function (method) {
    //监听函数
    var obj = this.globalData
    Object.defineProperty(obj, 'gpuProc', {
      configurable: true,
      enumerable: true,
      set: function (value) {
        this._gpuProc = value;
        method(value);
      },

      get: function () {
        if (this._gpuProc == undefined) {
          return []
        } else {
          return this._gpuProc
        }
      }
    })
  },
  get_conn() {
    return socket
  },
  connect_server() {
    var _this = this;
    //创建websocket
    //正式地址使用wss
    wx.closeSocket({
      complete(res) {
        console.log('res:', res)
      }
    })
    socket = wx.connectSocket({
      url: 'wss://pris.ssdk.icu:8888/portal',
      success: res => {
        console.info('创建连接成功');

        //socketTaskId: 22
        // console.info(res);
      }
    });
    // console.info(socket);
    //事件监听
    socket.onOpen(function () {
      console.info('连接打开成功');
    });
    socket.onClose(function () {
      console.info('连接关闭成功');
    });
    socket.onError(function () {
      console.info('连接报错');
    });
    //服务器发送监听
    socket.onMessage(function (e) {
      var ori = JSON.parse(e.data);
      var Type = ori.Type
      if (Type == 0) {
        var data = ori.Data.Data
        var ip_list = getApp().globalData.machineIPs
        var info_list = getApp().globalData.gpuInfo
        // console.log(ip_list)
        if (info_list[data[0].Ip] == null) {
          ip_list.push([data[0].Ip, ori.Data.Name])
        }
        info_list[data[0].Ip] = data
        _this.globalData.machineIPs = ip_list.sort()
        _this.globalData.gpuInfo = info_list
      }
      if (Type == 2) {
        var _gpuProc = getApp().globalData.gpuProc
        var Ip = ori.Data[0].Ip
        var gpu_cnt = getApp().globalData.gpuInfo[Ip].length
        _gpuProc[Ip] = []
        for (let index = 0; index < gpu_cnt; index++) {
          _gpuProc[Ip].push(ori.Data.filter((proc) => {
            return (proc.Ip == Ip) && (proc.Id == index);
          }));
        }
        console.log(_gpuProc)
        _this.globalData.gpuProc = _gpuProc
      }
    });
  },
  globalData: {
    machineIPs: [],
    gpuInfo: {},
    gpuProc: {}
  },
  onLaunch() {
    this.connect_server()
  },

})