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
      var ori = JSON.parse(e.data);
      var Type = ori.Type
      if (Type == 0) {
        var data = ori.Data.Data
        var ip_list = getApp().globalData.machineIPs
        var info_list = getApp().globalData.gpuInfo
        // console.log(ip_list)
        if (info_list[data[0].Ip] == null) {
          var cur = Date.now()
          var past = data[0].Time
          var online = true
          if (cur - past > 20000) {
            online = false
          }
          ip_list.push([data[0].Ip, ori.Data.Name, online])
        } else {
          var ip_list = getApp().globalData.machineIPs
          for (let index = 0; index < ip_list.length; index++) {
            if (target == ip_list[index][0]) {
              ip_list[index][2] = true
              getApp().globalData.machineIPs = ip_list
              break
            }
          }
        }
        info_list[data[0].Ip] = data
        // console.log(Date.now())
        _this.globalData.machineIPs = ip_list.sort()
        _this.globalData.gpuInfo = info_list
      }
      if (Type == 1) {
        var target = ori.Data
        var ip_list = getApp().globalData.machineIPs
        for (let index = 0; index < ip_list.length; index++) {
          if (target == ip_list[index][0]) {
            ip_list[index][2] = false
            getApp().globalData.machineIPs = ip_list
            break
          }
        }
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
        // console.log(_gpuProc)
        _this.globalData.gpuProc = _gpuProc
      }
    });
  },
  globalData: {
    machineIPs: [],
    gpuInfo: {},
    gpuProc: {},
    userCode: ''
  },
  onLaunch() {
    var that = this;
    wx.login({
      success: res => {
        // 发送 res.code 到后端换取 openId, sessionKey, unionId
        // 后端访问请求获取用户openId
        // console.log(res.code);
        that.globalData.userCode = res.code;
      },
      fail: res => {
        // 登录失败
        console.log("登录失败！");
      }
    })
    this.connect_server()
  },

})