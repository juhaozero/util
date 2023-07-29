package main

import (
	"fmt"
	"time"
	"util/etcd"
)

func main() {
	// common.GetRandom(1, 2)

	ed, err := etcd.NewEtcd([]string{"127.0.0.1:2379"}, "", "")
	if err != nil {
		panic(err)
	}
	// data, err := ed.GetService("test", clientv3.WithPrefix())
	// if err != nil {
	// 	panic(err)
	// }
	reg, err := ed.NewServiceRegister(300)

	if err != nil {
		panic(err)
	}

	ed.ServiceRegister = reg

	ed.Put("test999", "1000")

	ed.PutService("test111", "10000")

	fmt.Println("reg", reg)
	time.Sleep(20 * time.Second)

}
