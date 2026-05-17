package util

import (
	"fmt"
	"strconv"

	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

func mustGetQuery(ctx web.Context, key string) (string, error) {
	str, ok := ctx.Query(key)
	if !ok || str == "" {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	return str, nil
}

func MustGetWebQueryString(ctx web.Context, key string) (string, error) {
	return mustGetQuery(ctx, key)
}

func MustGetWebQueryInt32(ctx web.Context, key string) (int32, error) {
	str, err := mustGetQuery(ctx, key)
	if err != nil {
		return 0, err
	}
	value, parseErr := strconv.ParseInt(str, 10, 32)
	if parseErr != nil {
		return 0, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return int32(value), nil
}

func MustGetWebQueryInt64(ctx web.Context, key string) (int64, error) {
	str, err := mustGetQuery(ctx, key)
	if err != nil {
		return 0, err
	}
	value, parseErr := strconv.ParseInt(str, 10, 64)
	if parseErr != nil {
		return 0, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func MustGetWebQueryInt(ctx web.Context, key string) (int, error) {
	str, err := mustGetQuery(ctx, key)
	if err != nil {
		return 0, err
	}
	value, parseErr := strconv.Atoi(str)
	if parseErr != nil {
		return 0, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func MustGetWebQueryBool(ctx web.Context, key string) (bool, error) {
	str, err := mustGetQuery(ctx, key)
	if err != nil {
		return false, err
	}
	value, parseErr := strconv.ParseBool(str)
	if parseErr != nil {
		return false, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func GetWebQueryBool(ctx web.Context, key string, defaultValue bool) (bool, error) {
	str, ok := ctx.Query(key)
	if !ok {
		return defaultValue, nil
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		return false, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func ParamWebString(ctx web.Context, key string) (string, error) {
	str := ctx.Param(key)
	if str == "" {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	return str, nil
}

func ParamWebInt32(ctx web.Context, key string) (int32, error) {
	str, err := ParamWebString(ctx, key)
	if err != nil {
		return 0, err
	}
	value, parseErr := strconv.ParseInt(str, 10, 32)
	if parseErr != nil {
		return 0, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return int32(value), nil
}

func ParamWebInt64(ctx web.Context, key string) (int64, error) {
	str, err := ParamWebString(ctx, key)
	if err != nil {
		return 0, err
	}
	value, parseErr := strconv.ParseInt(str, 10, 64)
	if parseErr != nil {
		return 0, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func ParamWebBool(ctx web.Context, key string) (bool, error) {
	str, err := ParamWebString(ctx, key)
	if err != nil {
		return false, err
	}
	value, parseErr := strconv.ParseBool(str)
	if parseErr != nil {
		return false, xerr.WithStatus(parseErr, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}
