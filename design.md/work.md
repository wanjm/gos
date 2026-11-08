## 2026-11-04 11:48:00

- **Work item:** shared RPC HTTP client
- **Result:** prpc and srpc call RestfulClient.Exchange; non-2xx is still decoded as the old envelope
- **Docs:** technical.md RPC clients; guide.md 3.6.1

1. print all the filename to check why is there more file can be skipped;
2. print the struct name to check whether there is more struct can be skipped;
3. 完成缓存的代码生成;