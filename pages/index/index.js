// index.js
// 获取应用实例
const app = getApp()
var socket = {};
Page({
  data: {
    machineIPs: [],
    gpuInfo: {}
  },
  watchBack_gpuinfo: function (value) {
    //要执行的方法        
    this.setData({
      gpuInfo:value
    })
  },
  watchBack_machineIps: function (value) {
    //要执行的方法        
    this.setData({
      machineIPs:value
    })
  },
  go_detail: function (e) {
    var target=e.currentTarget.id
    wx.navigateTo({
      url: '/pages/detail/detail?ip='+target,
    })
    console.log('go detail',e.currentTarget.id)
  },
  // 事件处理函数
  onLoad(options) {
    // this.connect_server()
    let that = this;
    getApp().watch_gpuInfo(that.watchBack_gpuinfo.bind(that)) //注册监听
    getApp().watch_machineIPs(that.watchBack_machineIps.bind(that))
    console.log("页面onLoad")
  },
  onRefresh() {
    //在当前页面显示导航条加载动画
    wx.showNavigationBarLoading();
    //显示 loading 提示框。需主动调用 wx.hideLoading 才能关闭提示框
    wx.showLoading({
      title: '刷新中...',
    })
    getApp().connect_server()
    console.log("refresh")
    // this.connect_server();
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