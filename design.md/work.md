## 2026-11-05 19:51:00

- **Work item:** prpc header stays a developer function
- **Result:** removed the generated call to an interface method; `header=Func` is implemented outside gen
- **Docs:** technical.md RPC clients; guide.md 3.6.2

## 2026-11-05 19:35:00

- **Work item:** prpc header may call an interface method
- **Result:** `header=Method` generates a closure that calls that method; the method's own request skips the header to avoid recursion
- **Docs:** technical.md RPC clients; guide.md 3.6.2

## 2026-11-05 19:32:00

- **Work item:** prpc header function lives outside gen
- **Result:** `header=Func` on the prpc interface points at a same-package function; gen only wires it
- **Docs:** technical.md RPC clients; guide.md 3.6.2

## 2026-11-05 17:00:00

- **Work item:** prpc interface header method
- **Result:** `@gos header` on a prpc interface method is wired to `RestfulClient.header` and sent with the other methods
- **Docs:** technical.md RPC clients; guide.md 3.6.2

## 2026-11-04 11:48:00

- **Work item:** shared RPC HTTP client
- **Result:** prpc and srpc call RestfulClient.Exchange; non-2xx is still decoded as the old envelope
- **Docs:** technical.md RPC clients; guide.md 3.6.1

1. print all the filename to check why is there more file can be skipped;
2. print the struct name to check whether there is more struct can be skipped;
3. 完成缓存的代码生成;