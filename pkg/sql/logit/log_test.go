package logit

// We don't need to run this all the time, but leaving it commented
// in the file in case we need to test it again.

// func TestLog(t *testing.T) {
// 	batch := 1000
// 	t1 := time.Now()

// 	for i := 0; i < batch; i++ {
// 		Append("test", fmt.Sprintf("%d: Hello World", i))
// 	}

// 	t2 := time.Now()
// 	timeDiff := t2.Sub(t1)
// 	avg := timeDiff / time.Duration(batch)
// 	fmt.Printf("TestLog, Elapsed Time: %s, Avg per line: %s\n", timeDiff, avg)
// }
