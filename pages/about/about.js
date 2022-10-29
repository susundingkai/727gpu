// pages/about/about.js
Page({

  /**
   * 页面的初始数据
   */
  data: {
    list:[
      {
          pagePath:"/pages/index/index",
          text:"首页",
          iconPath:"/images/home-line.png",
          selectedIconPath:"/images/home-fill.png",
          iconSize:24
      },
      {
        pagePath:"/pages/about/about",
        text:"关于",
        iconPath:"/images/profile-line.png",
        selectedIconPath:"/images/profile-fill.png",
        iconSize:24
    }],
    show:false
  },

  /**
   * 生命周期函数--监听页面加载
   */
  onLoad(options) {

  },

  /**
   * 生命周期函数--监听页面初次渲染完成
   */
  onReady() {

  },

  /**
   * 生命周期函数--监听页面显示
   */
  onShow() {
    if(Date.now()%2==0){
      this.setData({
        show:true
      })
    }else{
      this.setData({
        show:false
      }) 
    }
  },

  /**
   * 生命周期函数--监听页面隐藏
   */
  onHide() {

  },

  /**
   * 生命周期函数--监听页面卸载
   */
  onUnload() {

  },

  /**
   * 页面相关事件处理函数--监听用户下拉动作
   */
  onPullDownRefresh() {

  },

  /**
   * 页面上拉触底事件的处理函数
   */
  onReachBottom() {

  },

  /**
   * 用户点击右上角分享
   */
  onShareAppMessage() {

  }
})