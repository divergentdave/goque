package goque

import "github.com/syndtr/goleveldb/leveldb/opt"

var writeOptionsFlush = opt.WriteOptions{Sync: true}
