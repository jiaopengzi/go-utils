//
// FilePath    : go-utils\convert_test.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 测试转换工具
//

package utils

import (
	"testing"
)

func TestIsWebvtt(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "基本的VTT内容",
			content: `WEBVTT

00:01.000 --> 00:04.000
- Never drink liquid nitrogen.

00:05.000 --> 00:09.000
- It will perforate your stomach.
- You could die.`,
			want: true,
		},
		{
			name:    "最简单的VTT内容",
			content: `WEBVTT `,
			want:    true,
		},
		{
			name:    "只有WEBVTT头部和标题",
			content: `WEBVTT - This file has no cues.`,
			want:    true,
		},
		{
			name: "有文本标题和cue的通用WebVTT",
			content: `WEBVTT - This file has cues.

14
00:01:14.815 --> 00:01:18.114
- What?
- Where are we now?

15
00:01:18.171 --> 00:01:20.991
- This is big bat country.

16
00:01:21.058 --> 00:01:23.868
- [ Bats Screeching ]
- They won't get in your hair. They're after the bugs.`,
			want: true,
		},
		{
			name: "通用注释用法",
			content: `WEBVTT - Translation of that film I like

NOTE
This translation was done by Kyle so that
some friends can watch it with their parents.

1
00:02:15.000 --> 00:02:20.000
- Ta en kopp varmt te.
- Det är inte varmt.

2
00:02:20.000 --> 00:02:25.000
- Har en kopp te.
- Det smakar som te.

NOTE This last line may not translate well.

3
00:02:25.000 --> 00:02:30.000
- Ta en kopp`,
			want: true,
		},
		{
			name: "WebVTT文件自身中定义样式",
			content: `WEBVTT

STYLE
::cue {
  background-image: linear-gradient(to bottom, dimgray, lightgray);
  color: papayawhip;
}
/* Style blocks cannot use blank lines nor "dash dash greater than" */

NOTE comment blocks can be used between style blocks.

STYLE
::cue(b) {
  color: peachpuff;
}

00:00:00.000 --> 00:00:10.000
- Hello <b>world</b>.

NOTE style blocks cannot appear after the first cue.`,
			want: true,
		},
		{
			name: "无任何VTT格式的随机文本",
			content: `Some random text
without any VTT format`,
			want: false,
		},
		{
			name:    "空字符串",
			content: ``,
			want:    false,
		},
		{
			name: "不正确的时间分割符号",
			content: `WEBVTT

00:01.000 -- 00:04.000
- Never drink liquid nitrogen.

00:05.000 -- 00:09.000
- It will perforate your stomach.
- You could die.`,
			want: false,
		},
		{
			name: "不正确的时间表达式",
			content: `WEBVTT

00:01.000 --> 00:04
- Never drink liquid nitrogen.

00:05 --> 00:09.000
- It will perforate your stomach.
- You could die.`,
			want: false,
		},
		{
			name: "中文VTT",
			content: `WEBVTT

1
00:00:00.000 --> 00:00:05.000
这是中文字幕

2
00:00:05.000 --> 00:01:10.000
我正在测试中文字幕2

3
00:01:10.000 --> 00:02:10.000
我正在测试中文字幕3

4
00:02:10.000 --> 00:03:10.000
我正在测试中文字幕4

5
00:03:10.000 --> 00:04:10.000
我正在测试中文字幕5

6
00:04:10.000 -- 00:05:10.000
我正在测试中文字幕6`,
			want: false,
		},
		{
			name: "空字幕",
			content: `WEBVTT

1
00:00:00.000 --> 00:00:05.000
这是中文字幕

2
00:00:05.000 --> 00:01:10.000


`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := IsWebvtt(tt.content); got != tt.want {
				t.Errorf("IsWebvtt() = %v, want: %v,err: %v", got, tt.want, err)
			}
		})
	}
}
