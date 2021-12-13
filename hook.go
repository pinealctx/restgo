package restgo

// BeforeHookFunc 请求前钩子函数
type BeforeHookFunc func(req IRequest)

// AfterHookFunc 请求后钩子函数
type AfterHookFunc func(req IRequest, rsp IResponse)
