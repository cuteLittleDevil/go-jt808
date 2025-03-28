package main

type zlmCode int

const (
	Exception   zlmCode = -400 //代码抛异常
	InvalidArgs zlmCode = -300 //参数不合法
	SqlFailed   zlmCode = -200 //sql执行失败
	AuthFailed  zlmCode = -100 //鉴权失败
	OtherFailed zlmCode = -1   //业务代码执行失败，
	Success     zlmCode = 0    //执行成功
)

func (z zlmCode) String() string {
	switch z {
	case Exception:
		return "代码抛异常"
	case InvalidArgs:
		return "参数不合法"
	case SqlFailed:
		return "sql执行失败"
	case AuthFailed:
		return "鉴权失败"
	case OtherFailed:
		return "业务代码执行失败"
	default:
		return "执行成功"
	}
}
