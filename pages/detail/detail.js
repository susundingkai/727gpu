// pages/detail/detail.js
var ip
Page({
  data: {
    ip: '',
    gpuProc: [],
    popup: false,
    gpuid: 0,
    slide: 20
  },

  watchBack_gpuproc: function (value) {
    //要执行的方法        
    this.setData({
      gpuProc: getApp().globalData.gpuProc[ip]
    })
    //隐藏loading 提示框
    wx.hideLoading();
    //隐藏导航条加载动画
    wx.hideNavigationBarLoading();
    //停止下拉刷新
    wx.stopPullDownRefresh();
  },
  sel:function(e){
    this.setData({
      gpuid:parseInt(e.detail.currentKey)
    })
  },
  load_data(target) {
    wx.showLoading({
      title: '刷新中...',
    })
    wx.sendSocketMessage({
      data: JSON.stringify({
        Type: 2,
        Data: [{
          Target: target
        }]
      })
    })
  },
  sub: function (e) {
    this.setData({
      popup: !this.data.popup
    })
    wx.lin.initValidateForm(this)
  },
  submit: function (event) {
    var sel = this.data.gpuid
    var memTH=this.data.slide
    wx.requestSubscribeMessage({
      tmplIds: ['EqKN5V8NPPpMtt5bsrfe52TC5zJGI2dEfR1o8xKlNm0', '5rFW4do6BBssNT_hRiq-gffEK0IXd4NkcF0QIqh8BtU'],
      success(res) {
        console.log(res);
        wx.sendSocketMessage({
          data: JSON.stringify({
            Type: 3,
            Data: [{
              Target: ip,
              Id: sel,
              Code: getApp().globalData.userCode,
              MemTH: memTH
            }]
          })
        })
      },
      fail(res) {
        console.log("subscribe failed:", res);
      }
    })
  },
  changeMemTh: function (e) {
    this.setData({
      slide: e.detail.value
    })
  },
  onLoad(options) {
    wx.lin.initValidateForm(this)
    var that = this
    ip = options.ip
    this.setData({
      ip: ip
    })
    console.log(options.ip)
    getApp().watch_gpuProc(that.watchBack_gpuproc.bind(that)) //注册监听
    that.load_data(ip)
  },


  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {
    this.load_data(ip)
  },

})