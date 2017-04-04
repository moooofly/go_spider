// Copyright 2014 Hu Cong. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scheduler

import (
    "container/list"
    "crypto/md5"
    "github.com/moooofly/go_spider/core/common/request"
    "sync"
    //"fmt"
)

// 实现 queue 类型 scheduler
type QueueScheduler struct {
    locker *sync.Mutex
    rm     bool  // 是否要求去重
    rmKey  map[[md5.Size]byte]*list.Element // 用于快速判定指定 URL 是否已经被保存到 queue 并提供去重功能
    queue  *list.List  // 实现用于保存 URl(element) 的 queue
}

// rmDuplicate => 是否要求去重
func NewQueueScheduler(rmDuplicate bool) *QueueScheduler {
    queue := list.New()
    rmKey := make(map[[md5.Size]byte]*list.Element)
    locker := new(sync.Mutex)
    return &QueueScheduler{rm: rmDuplicate, queue: queue, rmKey: rmKey, locker: locker}
}

func (this *QueueScheduler) Push(requ *request.Request) {
    this.locker.Lock()
    var key [md5.Size]byte
    // 如果 rm 为 ture ，则表示要求去重
    if this.rm {
        key = md5.Sum([]byte(requ.GetUrl()))
        // 通过查看 map 中是否存在相应的 key 来判定 URL 是否已存在
        if _, ok := this.rmKey[key]; ok {
            this.locker.Unlock()
            return
        }
    }
    // 执行到这里的情况:
    // 1. rm 为 false ，即不要求去重；
    // 2. rm 为 true ，但此时 map 中没有找到 URL 对应的 key

    // 插入 queue 最后，此时可能的情况为：
    // 1. 若 rm 为 false ，则 queue 中可能存在具有相同 URL 的 request
    // 2. 若 rm 为 true ，则 queue 中只会为一个 URL 保存一个 request
    e := this.queue.PushBack(requ)
    if this.rm {
        this.rmKey[key] = e
    }
    this.locker.Unlock()
}

// NOTE: 名字变更为 Pop 不是更好么？
func (this *QueueScheduler) Poll() *request.Request {
    this.locker.Lock()
    if this.queue.Len() <= 0 { // NOTE: should be '=' enough?
        this.locker.Unlock()
        return nil
    }
    // 从 queue 的 head 处获取（不删除） element
    e := this.queue.Front()
    requ := e.Value.(*request.Request)
    key := md5.Sum([]byte(requ.GetUrl()))
    // 从 queue 中删除 element
    this.queue.Remove(e)
    if this.rm {
        delete(this.rmKey, key)
    }
    this.locker.Unlock()
    return requ
}

func (this *QueueScheduler) Count() int {
    this.locker.Lock()
    len := this.queue.Len()
    this.locker.Unlock()
    return len
}
