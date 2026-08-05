## 2026-08-03 15:25:59

1. in yxt_data, we find filter in itself, also we have filter in plaso-common-go, they have same name, and same url, so i plan to add code gos, to ignore filter from external package when name and url are the same;
2. In yxt_data, local filters and plaso-common-go filters share the same names and URLs. Update gos so that when an external package defines a filter with the same name and URL as a local one, the external filter is ignored.

## 2026-08-02 07:13:32

1. i plan to add restrpc to interface, which is similar to srpc; but when calling restfull http request, error is localerror or object returned from server when http code means failed; please modify code of gos to add support of this;
2. I plan to add `restrpc` for interfaces, similar to `srpc`. For RESTful HTTP calls, errors should be either a local error or the object returned by the server when the HTTP status indicates failure. Please update gos to support this.

## 2026-08-02 07:08:58

1. in guide.md show me example for srpc interface;
2. In guide.md, add an example for the SRPC interface;
