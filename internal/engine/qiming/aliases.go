package qiming

import "liki-engine/internal/engine/ganzhi"

type Wuxing = ganzhi.Wuxing

func wuxingFromChinese(name string) Wuxing { wx, err := ganzhi.ParseWuxing(name); if err != nil { return 0 }; return wx }
