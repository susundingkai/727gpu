// index.js
// 获取应用实例
const app = getApp()
var socket = {};
Page({
  data: {
    machineIPs: [],
    gpuInfo: {}
  },
  connect_server() {
    var _this = this;
    //创建websocket
    //正式地址使用wss
    wx.closeSocket({complete(res){
      console.log('res:', res)
    }})
    socket = wx.connectSocket({
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
    });
  },
  // 事件处理函数
  onLoad(options) {
    this.connect_server()
  },
  onRefresh() {
    //在当前页面显示导航条加载动画
    wx.showNavigationBarLoading();
    //显示 loading 提示框。需主动调用 wx.hideLoading 才能关闭提示框
    wx.showLoading({
      title: '刷新中...',
    })
    console.log("refresh")
    this.connect_server();
    //隐藏loading 提示框
    wx.hideLoading();
    //隐藏导航条加载动画
    wx.hideNavigationBarLoading();
    //停止下拉刷新
    wx.stopPullDownRefresh();
  },
  onPullDownRefresh: function () {
    //调用刷新时将执行的方法
    this.onRefresh();
  }
})