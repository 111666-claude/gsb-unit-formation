# unit-formation

阵型站位：把槽位换算成本地坐标，按朝向旋转后加上队长位置，只用 Go 标准库。

```
go test ./...
go run ./cmd/unit-formation --sample spacing
go run ./cmd/unit-formation --sample shuffle
```

## 口径（README 为准）

- **站位由 Slot 决定**：第 i 个槽位的本地坐标是
  `(Slot % Columns) * Spacing` 与 `(Slot / Columns) * Spacing`，行号向下取整，
  再按 `headingDeg` 旋转加上 `origin`。同排相邻单位的距离必须等于 `Spacing`。
  `--sample spacing`（Columns 3、Spacing 1.5、一排三个单位）按口径是
  `u0.x=0.00 u1.x=1.50 u2.x=3.00`，现在是 `u0.x=0.00 u1.x=1.00 u2.x=2.00`，
  间距被压成 1 米，单位互相重叠。
- **与输入顺序无关**：位置只由 `Slot` 决定，列表顺序打乱结果不变。
  `--sample shuffle`（列表里 u1 排在 u0 前面）按口径是 `u0.x=0.00 u1.x=1.50`，
  现在是 `u0.x=1.00 u1.x=0.00`——两个人的位置跟着列表顺序换了个儿。
- 不变量：整队平移 `origin` 时每个单位的位移都等于平移量；旋转 `headingDeg` 时任意两个
  单位之间的距离不变；同一个 `Slot` 只会出现在一个位置上。
- 规模：1000 个单位一次 O(n) 排完，内存 O(n)；不许两两比较距离，也不许为了排位置先做排序。

## 现在的行为

- `Layout` 的行列号来自 `index % Columns` 与 `index / Columns`，`Slot` 字段被完全忽略。
- 本地坐标没有乘 `Spacing`，间距实际是 1 米，与配置无关。

## 输出

```
u0.x=0.00 u1.x=1.50 u2.x=3.00
u0.x=0.00 u1.x=1.50
```

## 目录

```
formation.go            阵型站位
cmd/unit-formation      命令行入口
formation_test.go       go test 用例
```
