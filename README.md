# unit-formation

阵型站位：按槽位算本地坐标、避让障碍物、同排顺移找空位，只用 Go 标准库。

```
go test ./...
go run ./cmd/unit-formation --sample spacing
go run ./cmd/unit-formation --sample shuffle
go run ./cmd/unit-formation --sample blocked
```

## 口径（README 为准）

- **槽位决定站位**：第 `Slot` 个槽位的本地坐标是 `(Slot % Columns) * Spacing` 与
  `(Slot / Columns) * Spacing`（行号向下取整），按 `headingDeg` 旋转后加上 `origin`。
  同排相邻单位的距离必须等于 `Spacing`。
- **避让障碍物**：候选位置被障碍物挡住时，在同一排内顺移到下一个槽位；
  这一排走完就换下一排的第一列。落位顺序按 `Slot` 升序，与输入顺序无关。
- **不许重叠**：本次布局里已经被占用的槽位同样要顺移，任意两个单位的位置都不能相同；
  找不到空位的单位放到 `origin`（这类单位数量要能被观测到）。
- **与输入顺序无关**：同一个 `Slot` 只出现在一个位置上，列表顺序打乱结果不变。
- **查询代价**：`Checks()` 是障碍查询次数，每单位最多探测常数个候选格，
  障碍检测要建空间索引，不许对每个候选位置线性扫全部障碍物。
- **规模**：2000 个单位、20 万障碍物、每帧一次布局，O(n log n)，内存 O(障碍物数)。

## 输出契约（不改格式）

```
u0=0.00,0.00 u1=1.50,0.00 u2=3.00,0.00
u0=0.00,0.00 u1=1.50,0.00
u1=3.00,0.00 u2=0.00,1.50 checks<=12
```

## 常量

```
Columns = 3
Spacing = 1.5
```

## 目录

```
formation.go            阵型站位
cmd/unit-formation      命令行入口
formation_test.go       go test 用例
```
