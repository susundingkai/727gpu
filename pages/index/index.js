// index.js
// 获取应用实例
const app = getApp()

Page({
  data: {
    machineIPs: [],
    gpuInfo: {}
  },
  // 事件处理函数
  onLoad(options) {
    var _this = this;
    //创建websocket
    //正式地址使用wss
    var socket = wx.connectSocket({
      url: 'wss://pris.ssdk.icu/portal',
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

      var data = JSON.parse(e.data);
      var _info = _this.data.gpuInfo
      data = data.Data.Data
      if (_info[data[0].Ip] == null) {
        _this.data.machineIPs.push(data[0].Ip)
      }
      _info[data[0].Ip] = data
      console.log(_info)
      _this.setData({
        machineIPs: _this.data.machineIPs,
        gpuInfo: _info
      })
      // var list = _this.data.result;
      // list = list.concat([data]);
      // _this.setData({
      //   result: list
      // });
    });
  }
})