本篇介绍：@gos enum 自动定义与关联，前端自动显示

前面第 6、7 篇分别打通了 Flutter / TypeScript 的 API → 前端 HTTP 调用与结构体生成。本篇继续：后端用 const 定义枚举，schema 字段关联枚举后，前端自动得到强类型 enum、`fromValue` 解析、以及 `text()` 中文展示，表格一行即可显示状态文案。

示例来自 yxt_data（定义）+ boss（展示）。

---

Go Web 开发提速 8（gos）：@gos enum 自动定义与关联 —— 后端一份常量，前端直接显示文案

本文将生成适配知乎风格的简短亮点摘要，放在文章开头：

---

# Go Web 开发提速 8（gos）：@gos enum 自动定义与关联 —— 后端一份常量，前端直接显示文案

### 本文亮点

✅ 一份 Go const：`@gos enum` 定义取值 + 展示文案，运行 gos 后自动生成 Dart / TS 枚举  
✅ Schema 字段关联：`// @gos enum=Xxx` 后，前端字段类型从 `int` 升级为强类型 enum，JSON 自动 `fromValue` / `.value`  
✅ 前端直接显示：`item.status.text()` 即可出中文，告别 Map 硬编码与前后端文案漂移

本系列持续更新，前七篇回顾：

1. [Go Web 开发提速：基于 Spring 式注释方案，用 gos 自动生成运行代码](https://zhuanlan.zhihu.com/p/1937905040842004437)
2. [Go Web 开发提速 (gos)：Servlet 注解与参数解析全指南 —— 从定义到落地](https://zhuanlan.zhihu.com/p/1937994788919019061)
3. [Go Web 开发提速 3（gos）：Filter 实战与变量注入 —— 通用逻辑复用与依赖解耦](https://zhuanlan.zhihu.com/p/1942992392115446822)
4. [Go Web 开发提速 4（gos）：自动生成代码实战解析，破除 Spring 注入误解](https://zhuanlan.zhihu.com/p/1952837653339828295)
5. [Go Web 开发提速 5（gos）：数据库代码全自动生成 —— 多库统一 + 零硬编码 + 极致复用](https://zhuanlan.zhihu.com/p/1994878858147686183)
6. [Go Web 开发提速 6（gos）：API 前后端联动生成 —— 一键打通全栈，告别接口对接成本](https://zhuanlan.zhihu.com/p/2027513197754589261)
7. Go Web 开发提速 7（gos）：TypeScript HTTP Client 自动生成 —— Web / 小程序同样一键联动（见 `design.md/desc7.md`；发布后补知乎链接）

在第 6、7 篇中，我们已经打通「Go API 定义 → Flutter / TypeScript 调用代码一键生成」。本篇解决另一个高频痛点：**状态码 / 类型码如何在前后端统一取值、统一文案、统一展示**。

传统做法往往是：后端写一份 const，前端再手写一份 Map，表格里 `status == 1 ? '待处理' : ...`。一旦多一个状态，两边同时改，还容易漏。gos 的 `@gos enum` 把这条链路也自动化了。

---

## 一、效果先看：表格里一行文案

后端定义（yxt_data）：

```go
// @gos enum=StudyBoardStatus;
const (
	StudyBoardStatusWaittingWord = iota + 1 // 等待字幕文件生成
	StudyBoardStatusWordReady               // 字幕文件生成完成
	// ...
)
```

Schema 字段关联：

```go
// @gos enum=StudyBoardStatus;
Status int32 `json:"status"`
```

Boss 前端表格（生成代码）直接：

```dart
DataCell(Text(item.status.text())),
```

`item.status` 已是 `StudyBoardStatus` 枚举，不再是裸 `int`；`text()` 返回「等待字幕文件生成」这类展示文案。前后端共用同一份取值与文案来源。

---

## 二、整体链路（三步）

1. **定义**：在 Go 的 `const (` 块上写 `// @gos enum=EnumName`，成员行注释或 `displayWord` 作为展示文案；
2. **关联**：在 API schema 字段上方写 `// @gos enum=EnumName`（必须写在字段 doc 注释，不能只写在行尾 `//`）；
3. **生成**：运行 `gos`，产出：
   - `enum.gen.dart`（或 `enum.gen.ts`）：枚举 + `value` + `displayWord` + `text()` / `xxxText` + `fromValue`；
   - `schema.gen.dart` / `schema.gen.ts`：对应字段类型变为该 enum。

配置复用第 6、7 篇的 `FlutterPath` / `TsPath`。

---

## 三、定义枚举：`@gos enum`

### 1. 基本写法（推荐显式命名）

以课程类型为例（`yxt_data/business/package/entity/mysql/package_basic/const.go`）：

```go
// @gos enum=PackageType;
const (
	// @gos displayWord="普通课程";
	PackageType_Normal = 1 //普通课程
	PackageType_Combo  = 2 //组合课程
)
```

要点：

| 项 | 说明 |
|----|------|
| `enum=PackageType` | 枚举在全局注册的名字；前端 Dart/TS 类型名与此一致 |
| 成员名 | `EnumName_Member` 或 `EnumNameMember`；下划线后（或前缀截断后）变成前端成员名，如 `normal` / `combo` |
| 展示文案 | 优先 `// @gos displayWord="..."`；没有则用行尾注释 |
| 取值类型 | 支持 `int`（含 `int8`/`iota`）、`string`、`bool` |

### 2. 省略名字时自动推断

```go
// @gos enum;
const (
	PackageType_Normal = 1
	PackageType_Combo  = 2
)
```

gos 会按第一个常量名的 `_` 前缀推断枚举名为 `PackageType`。生产代码更建议写全 `enum=Name`，避免歧义。

### 3. iota / 字符串 / 布尔

```go
// @gos enum=Status;
const (
	Status_A = iota + 1 // 启用
	Status_B            // 停用
)

// @gos enum=AiAgentType;
const (
	AI_HWL_NORMAL = "hwlnormal" // 好未来通用
)

// @gos enum=Flag;
const (
	Flag_Yes = true  // 是
	Flag_No  = false // 否
)
```

同一 const 块内取值类型必须一致（不能 int 和 string 混用）。

---

## 四、关联到 API 字段

仅有 enum 定义还不够：要让某个 JSON 字段在前端变成枚举类型，需要在 **schema 字段的上方注释** 关联：

```go
// yxt_data/business/package/schema/course.n.go
type NewCourseInfo struct {
	// ...
	// @gos enum=PackageType
	Type int8 `json:"type"` // 课程类型：1-普通课程 2-组合课程
}
```

```go
// yxt_data/business/ai/schema/study_board_resource.n.go
type AiStudyBoardResourceInfo struct {
	// @gos enum=StudyBoardStatus;
	Status int32 `json:"status"`
}
```

注意：

1. **`@gos enum=...` 写在字段 doc（上一行）**，写在行尾 `json` 后面的 `//` 里不会被识别；
2. Go 侧字段类型仍是 `int8`/`int32` 等——**服务端传输与存储不变**；变化发生在前端生成物；
3. 枚举名必须与 `@gos enum=Xxx` 定义侧一致，否则 gos 会 WARNING 并跳过关联。

---

## 五、前端生成了什么？

### 1. `enum.gen.dart`（全量枚举）

运行 gos 后，Boss 的 `lib/data/http/enum.gen.dart` 中会有类似：

```dart
enum PackageType {
  normal(1, '普通课程'),
  combo(2, '组合课程');

  const PackageType(this.value, this.displayWord);
  final int value;
  final String displayWord;
  String text() => displayWord;

  static PackageType? fromValue(int? v) {
    if (v == null) return null;
    for (final e in values) {
      if (e.value == v) return e;
    }
    return null;
  }
}
```

每个成员携带：

- `value`：与后端 const 一致的编码值；
- `displayWord`：展示文案；
- `text()`：给 UI 用的快捷方法；
- `fromValue`：JSON 反序列化入口。

### 2. `schema.gen.dart`（字段类型升级）

关联后，`NewCourseInfo.type` 不再是 `int`：

```dart
class NewCourseInfo extends JSONParameter {
  /// 课程类型：1-普通课程 2-组合课程
  PackageType type;

  NewCourseInfo({
    this.type = PackageType.normal,
    // ...
  });

  factory NewCourseInfo.fromJson(Map<String, dynamic> json) {
    return NewCourseInfo(
      type: PackageType.fromValue((json['type'] as num?)?.toInt())
          ?? PackageType.normal,
      // ...
    );
  }

  Map<String, dynamic> toJson() {
    return {
      "type": type.value,
      // ...
    };
  }
}
```

序列化链路：

- 入：JSON 数字 / 字符串 → `Enum.fromValue(...)` → 强类型枚举；
- 出：枚举 → `.value` → 与后端一致的原始值。

默认值取枚举的**第一个成员**。

### 3. UI 直接显示

Boss 列表页生成代码（如 `study_board_resource.g.dart`、`course_task_tab.g.dart`）：

```dart
DataCell(Text(item.status.text())),
DataCell(Text(item.type.text())),
DataCell(Text(item.sourceType.text())),
```

业务侧也可以直接比较枚举，而不是魔法数字：

```dart
if (item.type != PackageTaskType.quickLive &&
    item.type != PackageTaskType.classLive) {
  // ...
}
```

---

## 六、TypeScript 同步生成

第 7 篇已说明 TS HTTP Client 的生成方式。对 enum 而言：同一份 Go 定义会同时产出 `enum.gen.ts`（`export enum` / const 对象 + `packageTypeText(v)` 文案函数），并在 `schema.gen.ts` 字段类型上关联。Flutter 用 `text()`，TS 用 `xxxText(v)`，文案同源。

---

## 七、和「手写 Map / AI 随手写文案」比，为什么还要框架？

1. **单一真相源**：取值与文案只维护在 Go const；改一处，重新跑 gos，前后端同步；
2. **类型安全**：前端是 enum，不是 `int` + 注释；错误状态在编译期暴露；
3. **稳定可复现**：生成结构固定，不依赖 AI 每次「临场发挥」写一套 Map；
4. **AI 更省事**：告诉 AI「给 status 挂上 StudyBoardStatus」，它补注解 + 跑 gos 即可，不必手写两端文案表。

---

## 八、实战 checklist（照抄即可）

1. 在 entity / 独立 const 文件中写：

```go
// @gos enum=YourStatus;
const (
	YourStatus_Pending = 1 // 待处理
	YourStatus_Done    = 2 // 已完成
)
```

2. 在 schema 响应字段上方关联：

```go
// @gos enum=YourStatus
Status int8 `json:"status"`
```

3. 确认 `project.public.toml` 已配置 `FlutterPath` 和/或 `TsPath`；
4. 运行 `gos`；
5. 前端使用：`item.status.text()`（Dart）或 `studyBoardStatusText(item.status)`（TS）。

---

## 九、常见注意点

1. **字段注解位置**：必须写在字段上方的 doc 注释；行尾 `// @gos enum=...` 无效；
2. **枚举名全局唯一**：重复注册会 WARNING 并保留先解析到的定义；
3. **displayWord vs 行注释**：需要与行尾注释不同的展示文案时，用 `displayWord`；
4. **未关联的字段**：即使定义了 enum，schema 未写 `enum=`，前端仍是原始 `int`/`String`；
5. **Go 业务代码**：服务端继续用 const 数值比较即可；enum 注解主要服务「跨端生成与展示」。

---

## 十、总结与后续

至此，在第 6、7 篇「Flutter / TS API 调用联动」之上，补齐了 **枚举定义 → 字段关联 → 前端强类型与文案展示** 的闭环。

定义在 Go，显示在 Flutter/TS；表格、筛选、分支判断都能直接用生成的 enum，不再维护第二份状态字典。

后续可继续扩展：筛选下拉自动枚举选项、多语言 `displayKey`、以及与表常量配置（如 `table_constant_config`）的更深联动。

---

## 附录：草稿要点（写作素材）

- 痛点：前后端各维护一份状态 Map，易漂移
- 定义：`@gos enum=Name` on const block；成员 displayWord / 行注释
- 关联：schema 字段 `// @gos enum=Name`
- 生成：enum.gen.dart / enum.gen.ts + schema 字段类型替换
- 示例：PackageType、StudyBoardStatus、PackageTaskType（yxt_data + boss）
- UI：`item.status.text()`；比较：`PackageTaskType.quickLive`
- 支持 int/iota/string/bool；TS 同步生成 text helper
