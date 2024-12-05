package jt1078

import "testing"

func TestPTTypeString(t *testing.T) {
	tests := []struct {
		name string
		args uint8
		want PTType
	}{
		{
			name: "G711A",
			args: 6,
			want: PTG711A,
		},
		{
			name: "G711U",
			args: 7,
			want: PTG711U,
		},
		{
			name: "AAC",
			args: 19,
			want: PTAAC,
		},
		{
			name: "MP3",
			args: 25,
			want: PTMP3,
		},
		{
			name: "H264",
			args: 98,
			want: PTH264,
		},
		{
			name: "H265",
			args: 99,
			want: PTH265,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if PTType(tt.args).String() != tt.want.String() {
				t.Errorf("String() got = %v\n want %v", PTType(tt.args), tt.want)
			}
		})
	}
}

func TestDataTypeString(t *testing.T) {
	tests := []struct {
		name string
		args uint8
		want DataType
	}{
		{
			name: "视频I祯",
			args: 0,
			want: DataTypeI,
		},
		{
			name: "视频P帧",
			args: 1,
			want: DataTypeP,
		},
		{
			name: "视频B帧",
			args: 2,
			want: DataTypeB,
		},
		{
			name: "音频帧",
			args: 3,
			want: DataTypeA,
		},
		{
			name: "透传数据",
			args: 4,
			want: DataTypePenetrate,
		},
		{
			name: "未知类型",
			args: 5,
			want: DataType(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if DataType(tt.args).String() != tt.want.String() {
				t.Errorf("String() got = %v\n want %v", DataType(tt.args), tt.want)
			}
		})
	}
}

func TestSubcontractTypeString(t *testing.T) {
	tests := []struct {
		name string
		args uint8
		want SubcontractType
	}{
		{
			name: "原子祯",
			args: 0,
			want: SubcontractTypeAtomic,
		},
		{
			name: "第一祯",
			args: 1,
			want: SubcontractTypeFirst,
		},
		{
			name: "最后祯",
			args: 2,
			want: SubcontractTypeLast,
		},
		{
			name: "中间祯",
			args: 3,
			want: SubcontractTypeMiddle,
		},
		{
			name: "未知类型",
			args: 5,
			want: SubcontractType(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if SubcontractType(tt.args).String() != tt.want.String() {
				t.Errorf("String() got = %v\n want %v", SubcontractType(tt.args), tt.want)
			}
		})
	}
}
