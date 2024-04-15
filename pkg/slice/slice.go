package slice

type IdxRange struct {
	Low, High int
}

func Partition(collectionLen, partitionSize int) chan IdxRange {
	c := make(chan IdxRange)
	if partitionSize <= 0 {
		close(c)
		return c
	}
	go func() {
		numFullPartitions := collectionLen / partitionSize
		var i int
		for ; i < numFullPartitions; i++ {
			c <- IdxRange{Low: i * partitionSize, High: (i + 1) * partitionSize}
		}

		if collectionLen%partitionSize != 0 { // left over
			c <- IdxRange{Low: i * partitionSize, High: collectionLen}
		}

		close(c)
	}()
	return c
}

func Contains[K comparable](items []K, item K) bool {
	if items == nil || len(items) == 0 {
		return false
	}
	for _, v := range items {
		if v == item {
			return true
		}
	}
	return false
}

func Sub[K comparable](items []K, start, end int) (ret []K) {
	if start >= end || len(items) == 0 {
		return
	}
	size := len(items)
	if start >= size {
		return
	}
	if end > size {
		end = size
	}
	return items[start:end]
}

func Map[K comparable](items []K, _func func(item K) K) (ret []K) {
	for _, _item := range items {
		ret = append(ret, _func(_item))
	}
	return
}

func Filter[K comparable](items []K, _func func(item K) bool) (ret []K) {
	for _, _item := range items {
		if _func(_item) {
			ret = append(ret, _item)
		}
	}
	return
}
