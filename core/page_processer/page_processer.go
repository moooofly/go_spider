// Copyright 2014 Hu Cong. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package page_processer

import (
    "github.com/moooofly/go_spider/core/common/page"
)

// 页面处理器
type PageProcesser interface {
    Process(p *page.Page)  // 页面处理
    Finish()               // 页面处理后清理工作
}
