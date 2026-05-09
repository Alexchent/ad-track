# vivo 广告投放服务端基本的对接流程
## 第一步获取授权 token

> 新建应用流程介绍：https://open-ad.vivo.com.cn/doc/index?id=162
>
> code获取流程：https://open-ad.vivo.com.cn/doc/index?id=390
>
> token获取流程：https://open-ad.vivo.com.cn/doc/index?id=395
> 

### [vivo 商业开放平台](https://open-ad.vivo.com.cn)
1. 登录mkt api后台，然后新建应用 (新建应用的时候，会有一个接口权限勾选，推广服务不要勾选，其他全选就行)
2. 应用审核通过后，进行授权
3. 授权成功再进行数据上传 

### [vivo 营销平台](https://ad.vivo.com.cn/marketing/home)
1. 登录商业开放平台，并完成协议签署，用营销平台主账号登录即可
2. 创建开发者应用-应用管理看板新建应用，填写应用名，应用介绍等信息，并勾选对应权限
3. 等应用审核通过后，再走账户授权流程（先获取code，再拿code去获取token，获取token的目的是数据回传需要用到） 
   - 获取`code`码： 授权页面地址进行重要字段替换；`clientId`(在审核通过的应用下面看）、`state`(需要授权的账户名）以及回调地址，url地址登录进去，勾选需要的接口，点击确认授权后，vivo商业开放平台返回code授权码（在网址栏），时效10分钟
   - 获取`token`：拿着获取到的code去获取token, (可以照着CURL里面获取token的请求地址进行code/client_id等字段替换，生成链接；打开这个链接就会生成accesstoken以及refreshtoken；)


这些都走完了，就可以进行正常的数据回传了，文档地址 https://open-ad.vivo.com.cn/doc/index?id=217


1）授权操作： 广告主登录需要授权的广告投放账户（需要进入到投放平台账户的首页），然后将授权URL手动粘贴至浏览器中访问，点击界面上的授权按钮完成授权操作，即完成了该账户对开发者应用的授权。

链接： https://open-ad.vivo.com.cn/OAuth?clientId={您的client_id}&state={开发者标识}&redirectUri={您的redirectUri}

示例： https://open-ad.vivo.com.cn/OAuth?clientId=20200828001&state=vivo01&redirectUri=https://ad.vivo.com