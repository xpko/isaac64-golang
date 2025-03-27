package isaac64

// -------------------------------------------------------------------
// 一些常量定义；这里和原始C代码约定保持一致
const (
	RANDSIZL = 8             // 通常ISAAC中建议 8
	RANDSIZ  = 1 << RANDSIZL // 2^8 = 256
	RANDMAX  = RANDSIZ - 1   // 255
)

// ISAAC64State 对应 struct isaac64_state
type ISAAC64State struct {
	mm      [RANDSIZ]uint64 // 状态表
	randrsl [RANDSIZ]uint64 // 存放一次批量生成的随机数
	aa      uint64
	bb      uint64
	cc      uint64

	// randcnt 用于指示当前 randrsl[] 里的可用位置。
	// 原始C里是 randcnt=RANDMAX 代表下一次取数时先减1。
	// 也可以通过 randcnt==0 检查是否需要再生成一次。
	randcnt int
}

// ind 原始C里的宏：ind(mm, x) = *(ub8*)((ub1*)(mm) + ((x) & ((RANDSIZ-1)<<3)))
// 解释：对 mm 做“按字节”的偏移，然后再取 64 位整型。
// 等价于在 Go 中： mm[( (x) & ((RANDSIZ-1)<<3)) >> 3]。
func ind(mm [RANDSIZ]uint64, x uint64) uint64 {
	// (RANDSIZ-1) << 3 = 255 << 3 = 2040
	// 因此 (x & 2040) >> 3 取出的是 x 低 8 位索引(每 8 字节为1个uint64元素)。
	return mm[(x&((RANDSIZ-1)<<3))>>3]
}

// mix 对应原始C里的宏 mix(a,b,c,d,e,f,g,h)
func mix(a, b, c, d, e, f, g, h uint64) (na, nb, nc, nd, ne, nf, ng, nh uint64) {
	a -= e
	f ^= (h >> 9)
	h += a
	b -= f
	g ^= (a << 9)
	a += b
	c -= g
	h ^= (b >> 23)
	b += c
	d -= h
	a ^= (c << 15)
	c += d
	e -= a
	b ^= (d >> 14)
	d += e
	f -= b
	c ^= (e << 20)
	e += f
	g -= c
	d ^= (f >> 17)
	f += g
	h -= d
	e ^= (g << 14)
	g += h
	return a, b, c, d, e, f, g, h
}

// isaac64Generate 对应原C函数 isaac64_generate
// 每次调用会填充 rng.randrsl[]，生成 RANDSIZ 个新的随机数。
func isaac64Generate(rng *ISAAC64State) {
	var x, y uint64
	a := rng.aa
	b := rng.bb + (rng.cc + 1) // b = rng->bb + ++rng->cc
	rng.cc++

	// 前半段：m=[0..127], m2=[128..255]
	i := 0
	i2 := RANDSIZ / 2 // 128
	rIdx := 0         // randrsl的下标
	for i < RANDSIZ/2 {
		// step1
		x = rng.mm[i]
		a = (^(a ^ (a << 21))) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++

		// step2
		x = rng.mm[i]
		a = (a ^ (a >> 5)) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++

		// step3
		x = rng.mm[i]
		a = (a ^ (a << 12)) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++

		// step4
		x = rng.mm[i]
		a = (a ^ (a >> 33)) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++
	}

	// 后半段：m=[128..255], m2=[0..127]
	i2 = 0
	for i < RANDSIZ {
		// step1
		x = rng.mm[i]
		a = (^(a ^ (a << 21))) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++

		// step2
		x = rng.mm[i]
		a = (a ^ (a >> 5)) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++

		// step3
		x = rng.mm[i]
		a = (a ^ (a << 12)) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++

		// step4
		x = rng.mm[i]
		a = (a ^ (a >> 33)) + rng.mm[i2]
		i2++
		y = ind(rng.mm, x) + a + b
		rng.mm[i] = y
		i++
		b = ind(rng.mm, y>>RANDSIZL) + x
		rng.randrsl[rIdx] = b
		rIdx++
	}

	rng.bb = b
	rng.aa = a
	// 准备下一次取随机数时使用
	rng.randcnt = RANDMAX
}

// isaac64Init 对应原C函数 isaac64_init
// 传入一个 seed（32位），初始化状态 rng，并生成初始的一组随机数
func (rng *ISAAC64State) Isaac64Init(seed uint64) {
	var a, b, c, d, e, f, g, h uint64
	// 经典魔数：the golden ratio
	a, b, c, d, e, f, g, h = 0x9e3779b97f4a7c13, 0x9e3779b97f4a7c13,
		0x9e3779b97f4a7c13, 0x9e3779b97f4a7c13,
		0x9e3779b97f4a7c13, 0x9e3779b97f4a7c13,
		0x9e3779b97f4a7c13, 0x9e3779b97f4a7c13

	// 初始化 rng 内部参数
	rng.aa = 0
	rng.bb = 0
	rng.cc = 0

	// randrsl 全部置 0
	for i := 0; i < RANDSIZ; i++ {
		rng.randrsl[i] = 0
	}
	// 这里只使用 seed 写入第一个位置，其余为 0
	rng.randrsl[0] = uint64(seed)

	// 先做4次搅乱
	for i := 0; i < 4; i++ {
		a, b, c, d, e, f, g, h = mix(a, b, c, d, e, f, g, h)
	}

	// 第一遍：把 randrsl[] 的数据注入到 mm[] 里
	for i := 0; i < RANDSIZ; i += 8 {
		a += rng.randrsl[i+0]
		b += rng.randrsl[i+1]
		c += rng.randrsl[i+2]
		d += rng.randrsl[i+3]
		e += rng.randrsl[i+4]
		f += rng.randrsl[i+5]
		g += rng.randrsl[i+6]
		h += rng.randrsl[i+7]
		a, b, c, d, e, f, g, h = mix(a, b, c, d, e, f, g, h)
		rng.mm[i+0] = a
		rng.mm[i+1] = b
		rng.mm[i+2] = c
		rng.mm[i+3] = d
		rng.mm[i+4] = e
		rng.mm[i+5] = f
		rng.mm[i+6] = g
		rng.mm[i+7] = h
	}

	// 第二遍：再用 mm[] 自身进行一次搅乱
	for i := 0; i < RANDSIZ; i += 8 {
		a += rng.mm[i+0]
		b += rng.mm[i+1]
		c += rng.mm[i+2]
		d += rng.mm[i+3]
		e += rng.mm[i+4]
		f += rng.mm[i+5]
		g += rng.mm[i+6]
		h += rng.mm[i+7]
		a, b, c, d, e, f, g, h = mix(a, b, c, d, e, f, g, h)
		rng.mm[i+0] = a
		rng.mm[i+1] = b
		rng.mm[i+2] = c
		rng.mm[i+3] = d
		rng.mm[i+4] = e
		rng.mm[i+5] = f
		rng.mm[i+6] = g
		rng.mm[i+7] = h
	}

	// 生成第一批随机数
	isaac64Generate(rng)
}

// 取一个 64 位随机数的简单方法
func (rng *ISAAC64State) Isaac64Rand() uint64 {
	// 如果已经用光，就再次生成
	if rng.randcnt < 0 || rng.randcnt >= RANDSIZ {
		// 理论上 randcnt=RANDMAX=255 时开始取数
		// 取完 256次后变成 -1，触发重新生成
		isaac64Generate(rng)
	}
	val := rng.randrsl[rng.randcnt]
	rng.randcnt--
	return val
}

func New() *ISAAC64State {
	return new(ISAAC64State)
}
