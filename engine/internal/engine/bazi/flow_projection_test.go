package bazi

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func TestFlowProjectionsNormalizeOffset(t *testing.T) {
	atZ, err := time.Parse(time.RFC3339Nano, "1990-05-20T04:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	atOffset, err := time.Parse(time.RFC3339Nano, "1990-05-20T12:00:00+08:00")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	dayunZ, err := ComputeDaYunAt(atZ.UTC(), loc, 116.407526, ganzhi.Male)
	if err != nil {
		t.Fatal(err)
	}
	dayunOffset, err := ComputeDaYunAt(atOffset.UTC(), loc, 116.407526, ganzhi.Male)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dayunZ, dayunOffset) {
		t.Fatal("offset serialization changed dayun")
	}

	liunianZ, err := ComputeLiuNianAt(atZ.UTC(), loc, 116.407526, 2026)
	if err != nil {
		t.Fatal(err)
	}
	liunianOffset, err := ComputeLiuNianAt(atOffset.UTC(), loc, 116.407526, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(liunianZ, liunianOffset) {
		t.Fatal("offset serialization changed liunian")
	}

	liuyueZ, err := ComputeLiuYueAt(atZ.UTC(), loc, 116.407526, 2026, 5)
	if err != nil {
		t.Fatal(err)
	}
	liuyueOffset, err := ComputeLiuYueAt(atOffset.UTC(), loc, 116.407526, 2026, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(liuyueZ, liuyueOffset) {
		t.Fatal("offset serialization changed liuyue")
	}

	liuriZ, err := ComputeLiuRiAt(atZ.UTC(), loc, 116.407526, 2026, 5, 20)
	if err != nil {
		t.Fatal(err)
	}
	liuriOffset, err := ComputeLiuRiAt(atOffset.UTC(), loc, 116.407526, 2026, 5, 20)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(liuriZ, liuriOffset) {
		t.Fatal("offset serialization changed liuri")
	}

	liushiZ, err := ComputeLiuShiAt(atZ.UTC(), loc, 116.407526, 2026, 5, 20, 12)
	if err != nil {
		t.Fatal(err)
	}
	liushiOffset, err := ComputeLiuShiAt(atOffset.UTC(), loc, 116.407526, 2026, 5, 20, 12)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(liushiZ, liushiOffset) {
		t.Fatal("offset serialization changed liushi")
	}

	xiaoyunZ := ComputeXiaoYunAt(atZ.UTC(), loc, 116.407526, 3)
	xiaoyunOffset := ComputeXiaoYunAt(atOffset.UTC(), loc, 116.407526, 3)
	if !reflect.DeepEqual(xiaoyunZ, xiaoyunOffset) {
		t.Fatal("offset serialization changed xiaoyun")
	}
}
