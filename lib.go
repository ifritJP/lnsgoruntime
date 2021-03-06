/*
MIT License

Copyright (c) 2018,2020 ifritJP

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package lnsgoruntime

import "fmt"
import "math"
import "reflect"
import "strconv"

type LnsInt = int
type LnsReal = float64
type LnsAny = interface{}

type LnsForm func( []LnsAny ) []LnsAny

var LnsNone interface{} = nil
var Lns_package_path string

type LnsEnv struct {
    valStack []LnsAny
    nilAccStack []LnsAny
    LuaVM *Lns_luaVM
}
var cur_LnsEnv *LnsEnv

func Lns_GetEnv () *LnsEnv {
    return cur_LnsEnv
}

type LnsThreadMtd interface {
    Loop()
}

type LnsThread struct {
    LnsEnv *LnsEnv
    FP LnsThreadMtd
}
func (self *LnsThread) InitLnsThread() {
    self.LnsEnv = createEnv()
}



/**
各モジュールを初期化する際に実行する関数。

import エラーを回避するため、
敢てランタイムの関数のどれか一つを呼んでいる。

実際の初期化関数は、 Lns_InitModOnce() で行なう。
*/
func Lns_InitMod() {
}

func createEnv() *LnsEnv {
    env := &LnsEnv{}
    env.valStack = []LnsAny{}
    env.nilAccStack = []LnsAny{}
    env.LuaVM = createVM()

    return env
}

/**
各モジュールを初期化する際に実行する関数。
*/
func Lns_InitModOnce() {
    cur_LnsEnv = createEnv()
    
    Lns_package_path = cur_LnsEnv.LuaVM.GetPackagePath()
}


func Lns_IsNil( val LnsAny ) bool {
    if val == nil {
        return true;
    }
    switch val.(type) {
    case LnsInt:
        return false;
    case LnsReal:
        return false;
    case bool:
        return false;
    case string:
        return false;
    default:
        value := reflect.ValueOf(val)
        if value.Kind() == reflect.Func {
            return false
        }
        return value.IsNil()
    }    
}

func Lns_type( val LnsAny ) string {
    if val == nil {
        return "nil";
    }
    switch val.(type) {
    case LnsInt:
        return "number";
    case LnsReal:
        return "number";
    case bool:
        return "boolean";
    case string:
        return "string";
    case *Lns_luaValue:
        switch val.(*Lns_luaValue).core.typeId {
        case cLUA_TFUNCTION:
            return "function"
        case cLUA_TTABLE:
            return "table"
        default:
            return fmt.Sprintf( "<notsupport>:%d, %d, %s",
                val.(*Lns_luaValue).core.typeId, cLUA_TTABLE,
                val.(*Lns_luaValue).core.typeId == cLUA_TTABLE)
        }
    default:
        value := reflect.ValueOf(val)
        if value.Kind() == reflect.Func {
            return "function";
        }
        return "table";
    }    
}

func Lns_tonumber( val string, base LnsAny ) LnsAny {
    if Lns_IsNil( base ) {
        if Str_startsWith( val, "0x" ) || Str_startsWith( val, "0X" ) {
            if dig, err := strconv.ParseInt( val[2:], 16, 64 ); err != nil {
                return nil
            } else {
                return LnsReal( dig )
            }
        }
        f, err := strconv.ParseFloat( val, 64)
        if err != nil {
            return nil
        }
        return f
    }
    if bs, ok := base.(LnsInt); ok {
        print( fmt.Sprint( "hoge: %s %s", val, bs ) );
        if dig, err := strconv.ParseInt( val, bs, 64 ); err != nil {
            return nil
        } else {
            return LnsReal( dig )
        }
    } else {
        return nil
    }
}

func Lns_require( val string ) LnsAny {
    panic( "not support" )
    return nil
}


func Lns_unwrap( val LnsAny ) LnsAny {
    if Lns_IsNil( val ) {
        panic( "unwrap nil" );
    }
    return val
}
func Lns_unwrapDefault( val, def LnsAny ) LnsAny {
    if Lns_IsNil( val ) {
        return def
    }
    return val
}

func Lns_cast2LnsInt( val LnsAny ) LnsAny {
    if _, ok := val.(LnsInt); ok {
        return val
    }
    return nil
}
func Lns_cast2LnsReal( val LnsAny ) LnsAny {
    if _, ok := val.(LnsReal); ok {
        return val
    }
    return nil
}
func Lns_cast2bool( val LnsAny ) LnsAny {
    if _, ok := val.(bool); ok {
        return val
    }
    return nil
}
func Lns_cast2string( val LnsAny ) LnsAny {
    if _, ok := val.(string); ok {
        return val
    }
    return nil
}

func Lns_forceCastInt( val LnsAny ) LnsInt {
    switch val.(type) {
    case LnsInt:
        return val.(LnsInt)
    case LnsReal:
        return LnsInt(val.(LnsReal))
    }
    panic( fmt.Sprintf( "illegal type -- %T", val ) );
}

func Lns_forceCastReal( val LnsAny ) LnsReal {
    switch val.(type) {
    case LnsInt:
        return LnsReal(val.(LnsInt))
    case LnsReal:
        return val.(LnsReal)
    }
    panic( fmt.Sprintf( "illegal type -- %T", val ) );
}

type Lns_ToObjParam struct {
    Func func ( obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny)
    Nilable bool
    Child []Lns_ToObjParam
}
type Lns_ToObj func ( obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny)

func Lns_ToStemSub(
    obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    return true, obj, nil
}
func Lns_ToIntSub(
    obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    if _, ok := obj.(LnsInt); ok {
        return true, obj, nil
    }
    return false, nil, "no int"
}
func Lns_ToRealSub( obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    if _, ok := obj.(LnsReal); ok {
        return true, obj, nil
    }
    return false, nil, "no real"
}
func Lns_ToBoolSub( obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    if _, ok := obj.(bool); ok {
        return true, obj, nil
    }
    return false, nil, "no bool"
}
func Lns_ToStrSub( obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    if _, ok := obj.(string); ok {
        return true, obj, nil
    }
    return false, nil, "no str"
}

type LnsAlgeVal interface {
    GetTxt() string
}

func Lns_getFromMulti( multi []LnsAny, index LnsInt ) LnsAny {
    if len( multi ) > index {
        return multi[ index ];
    }
    return nil;
}

func (self *LnsEnv) NilAccPush( obj interface{} ) bool {
    if Lns_IsNil( obj )  {
        return false
    }
    self.nilAccStack = append( self.nilAccStack, obj )
    return true
}

// func Lns_NilAccLast( obj interface{} ) bool {
//     self.nilAccStack = append( self.nilAccStack, obj )
//     return true
// }

func (self *LnsEnv) NilAccPop() LnsAny {
    obj := self.nilAccStack[ len( self.nilAccStack ) - 1 ]
    self.nilAccStack = self.nilAccStack[ : len( self.nilAccStack ) - 1 ]
    return obj
}

func (self *LnsEnv) NilAccFin( ret bool) LnsAny {
    if ret {
        return self.NilAccPop()
    }
    return nil
}

func Lns_NilAccCall0( self *LnsEnv, call func () ) bool {
    call()
    return self.NilAccPush( true )
}
func Lns_NilAccCall1( self *LnsEnv, call func () LnsAny ) bool {
    return self.NilAccPush( call() )
}
func Lns_NilAccCall2( self *LnsEnv, call func () (LnsAny,LnsAny) ) bool {
    return self.NilAccPush( Lns_2DDD( call() ) )
}
func Lns_NilAccFinCall2( ret LnsAny ) (LnsAny,LnsAny) {
    if Lns_IsNil( ret ) {
        return nil, nil
    }
    list := ret.([]LnsAny)
    return list[0], list[1]
}
func Lns_NilAccCall3( self *LnsEnv, call func () (LnsAny,LnsAny,LnsAny) ) bool {
    return self.NilAccPush( Lns_2DDD( call() ) )
}
func Lns_NilAccFinCall3( ret LnsAny ) (LnsAny,LnsAny,LnsAny) {
    if Lns_IsNil( ret ) {
        return nil, nil, nil
    }
    list := ret.([]LnsAny)
    return list[0], list[1], list[2]
}



func Lns_isCondTrue( stem LnsAny ) bool {
    if Lns_IsNil( stem ) {
        return false;
    }
    switch stem.(type) {
    case bool:
        return stem.(bool)
    default:
        return true
    }
}

func Lns_op_not( stem LnsAny ) bool {
    return !Lns_isCondTrue( stem );
}

/** 多値返却の先頭を返す
*/
func Lns_car( multi ...LnsAny ) LnsAny {
    if len( multi ) == 0 {
        return nil
    }
    if Lns_IsNil( multi[0] ) {
        return nil
    }
    if ddd, ok := multi[ 0 ].([]LnsAny); ok {
        return Lns_car( ddd... )
    }
    return multi[0]
}

func Lns_2DDD( multi ...LnsAny ) []LnsAny {
    if len( multi ) == 0 {
        return multi;
    }
    switch multi[ len( multi ) - 1 ].(type) {
    case []LnsAny:
        ddd := multi[ len( multi ) - 1 ].([]LnsAny)
        newMulti := multi[ :len( multi ) - 1 ];
        for _, val := range( ddd ) {
            newMulti = append( newMulti, val )
        }
        return newMulti
    }
    return multi
}

func Lns_print( multi []LnsAny ) {
    for index, val := range( multi ) {
        if index != 0 {
            fmt.Print( "\t" )
        }
        fmt.Print( Lns_ToString( val ) )
    }
    fmt.Print( "\n" )
}


func Lns_ToString( val LnsAny ) string {
    if Lns_IsNil( val ) {
        return "nil"
    }
    switch val.(type) {
    case LnsInt:
        return fmt.Sprintf( "%d", val )
    case LnsReal:
        if digit, frac := math.Modf( val.(LnsReal) ); frac == 0 {
            return fmt.Sprintf( "%g.0", digit )
        }
        return fmt.Sprintf( "%g", val )
    case bool:
        return fmt.Sprintf( "%v", val )
    case string:
        return val.(string);
    default:
        value := reflect.ValueOf(val)
        if value.Kind() == reflect.Func {
            return fmt.Sprintf( "function:%T", val )
        }
        return fmt.Sprintf( "table:%T", val )
    }
}

/**
 * スタックを一段上げる
 */
func (self *LnsEnv) IncStack() bool {
    self.valStack = append( self.valStack, nil )
    return false;
}

/**
 * 値 pVal をスタックの top にセットし、値 pVal の条件判定結果を返す
 *
 * スタックに lns_ddd_t は詰めない。
 * 呼び出し側で lns_ddd_t の先頭要素を指定すること。
 *
 * @param pVal スタックに詰む値
 * @return pVal の条件判定結果。 lns_isCondTrue()。
 */
func (self *LnsEnv) SetStackVal( val LnsAny ) bool {
    self.valStack[ len( self.valStack ) - 1 ] = val
    return Lns_isCondTrue( val )
}

/**
 * スタックから値を pop する。
 *
 * @return pop した値。
 */
func (self *LnsEnv) PopVal( dummy bool ) LnsAny {
    val := self.valStack[ len( self.valStack ) - 1 ]
    self.valStack = self.valStack[:len(self.valStack) - 1]
    return val;
}
