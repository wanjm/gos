# Flutter Code Generation Requirements

## Overview
Generate Flutter/Dart network client code and DTOs from Golang struct and function definitions annotated for the servlet framework.

## Input Source
- **Golang Structs**: Marked with `// @gos type=servlet` (or similar group tags).
- **Golang Methods**: Attached to the struct, marked with `// @gos url="/path"`.
- **Golang Types**: Request and Response structs used in the methods.

## Output Configuration
- **Directory**: `frontgen/flutter` (relative to project root).

## Generated Code Structure

### Imports
Generated files must include:
```dart
import 'package:http_method/http_method.dart';
import 'myclient.dart';
```

### DTOs (Data Transfer Objects)
For each Go struct used in the API:
1.  **Class Definition**: Extends `JSONParameter` (from `http_method`).
2.  **Fields**: `final` typed fields.
3.  **Constructor**: Named parameters with `required`.
4.  **Serialization**:
    -   `toJson()`: Returns `Map<String, dynamic>`.
    -   `fromJson(Map<String, dynamic> json)`: Factory constructor.
5.  **Null Safety & Defaults**:
    -   Strings: default to `""` if null (`json['key'] ?? ''`).
    -   Numbers: default to `0`/`0.0`.
    -   Lists: default to empty list `[]` if null.
    -   Handle nested object parsing.

### Network Service

#### Abstract Interface
Define an abstract class for the service:
```dart
abstract class HelloNetwork {
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data);
}
```

#### Implementation
Define the implementation class extending `BaseMethod` (from `http_method`):
```dart
class HelloNetworkImpl extends BaseMethod implements HelloNetwork {
  HelloNetworkImpl({required HttpClient client}) : super(client: client);

  @override
  Future<RespData<HelloResponse?>> sayHello(HelloRequest data) => getData(
    data: data,
    url: "/hello", // derived from Go annotation
    encodeDataFunction: (RespData resp) {
      // Parse logic
      resp.obj = HelloResponse.fromJson(resp.res);
    },
  );
}
```

### Global Instance
Generate a global variable instantiation relying on a global `client` from `myclient.dart`:
```dart
var helloNetwork = HelloNetworkImpl(client: client);
```

## Dependencies
- **package:http_method**: Provides `BaseMethod`, `RespData`, `JSONParameter`, `HttpClient`.
- **myclient.dart**: Expected to provide the global `client` instance.




