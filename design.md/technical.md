# gos technical

注解、生成物和目录见 **[guide.md](guide.md)**。

## RPC clients

prpc、srpc、restrpc 共用 `RestfulClient.Exchange`（生成在 `rpc_common.gen.go`）。Exchange 只负责 HTTP：组 URL、写 body、trace header、打日志，返回 status 和 body。

包络留在各自的 client：

| Client | 请求 | 响应 | status |
|--------|------|------|--------|
| prpc | JSON array | `{c,o}`，`o` 为 `[error, obj]` | 忽略，任何 status 都解码 body |
| srpc | JSON object（无参时 body 为 `null`） | `{code,msg,obj}` | 忽略 |
| restrpc | JSON object；GET/HEAD 为 query | body 即业务对象 | 非 2xx 返回 `*RestError` |
