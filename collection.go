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
import "sort"

type Lns_ToMap interface {
    ToMap() *LnsMap
}

type Lns_ToCollectionIF interface {
    ToCollection() LnsAny
}

func Lns_ToCollection( val LnsAny ) LnsAny {
    if Lns_IsNil( val ) {
        return nil
    }
    switch val.(type) {
    case LnsInt:
        return val
    case LnsReal:
        return val
    case bool:
        return val
    case string:
        return val
    case Lns_ToCollectionIF:
        return val.(Lns_ToCollectionIF).ToCollection()
    default:
        return val.(Lns_ToMap).ToMap()
    }
}

func Lns_FromStemGetAt( obj LnsAny, index LnsAny, nilAccess bool ) LnsAny {
    if nilAccess {
        if Lns_IsNil( obj ) {
            return nil
        }
    }
    if mapObj, ok := obj.(*LnsMap); ok {
        return mapObj.Items[ index ]
    }
    return obj.(*LnsList).Items[ index.(LnsInt) - 1 ]
}


// ======== list ========

const (
    LnsItemKindUnknown = 0
    LnsItemKindStem = 1
    LnsItemKindInt = 2
    LnsItemKindReal = 3
    LnsItemKindStr = 4
)

type LnsList struct {
    Items []LnsAny
    lnsItemKind int
}

func Lns_ToListSub(
    obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    itemParam := paramList[0]
    if val, ok := obj.(*LnsList); ok {
        list := make([]LnsAny, len(val.Items))
        for index, val := range( val.Items ) {
            success, conved, mess :=
                itemParam.Func( val, itemParam.Nilable, itemParam.Child )
            if !success {
                return false, nil, fmt.Sprintf( "%d:%s", index + 1, mess )
            }
            list[ index ] = conved
        }
        return true, NewLnsList( list ), nil
    } else if val, ok := obj.(*LnsMap); ok {
        // lua は list と map に明確な差がないので、
        // 本来 list のデータも map になる可能性があるため、
        // map からも処理できるようにする。
        list := make([]LnsAny, len(val.Items))
        for index := 1; index <= len( val.Items ); index++ {
            if val, ok := val.Items[ index ]; ok {
                success, conved, mess :=
                    itemParam.Func( val, itemParam.Nilable, itemParam.Child )
                if !success {
                    return false, nil, fmt.Sprintf( "%d:%s", index, mess )
                }
                list[ index - 1 ] = conved
            } else {
                return false, nil, fmt.Sprintf( "%d:%s", index, "no index" )
            }
        }
        return true, NewLnsList( list ), nil
    }
    return false, nil, "no list"
}

func (self *LnsList) ToCollection() LnsAny {
    list := make([]LnsAny, len(self.Items))
    for index, val := range (self.Items) {
        list[ index ] = Lns_ToCollection( val )
    }
    return NewLnsList( list )
}

type LnsComp func( val1, val2 LnsAny ) bool;

func (self *LnsList) Sort( kind int, comp LnsAny ) {
    if self.lnsItemKind == LnsItemKindUnknown || self.lnsItemKind == LnsItemKindStem {
        self.lnsItemKind = kind
    }
    if Lns_IsNil( comp ) {
        if self.lnsItemKind == LnsItemKindStem {
            hasInt := 0
            hasReal := 0
            hasStr := 0
            hasStem := false
            for _, val := range( self.Items ) {
                switch val.(type) {
                case LnsInt:
                    hasInt = 1
                case LnsReal:
                    hasReal = 1
                case string:
                    hasStr = 1
                default:
                    break
                }
            }
            if !hasStem && (hasInt + hasReal + hasStr) == 1 {
                if hasInt == 1 {
                    self.lnsItemKind = LnsItemKindInt
                } else if hasReal == 1 {
                    self.lnsItemKind = LnsItemKindReal
                } else if hasStr == 1 {
                    self.lnsItemKind = LnsItemKindStr
                }
            }
        }
        sort.Sort( self )
    } else {
        callback := comp.(LnsComp)
        sort.Slice(
            self.Items,
            func (idx1, idx2 int ) bool {
                return callback( self.Items[ idx1 ], self.Items[ idx2 ] )
            } )
    }
}

func (self *LnsList) Len() int {
    return len(self.Items)
}
func (self *LnsList) Less(idx1, idx2 int) bool {
    val1 := self.Items[idx1]
    val2 := self.Items[idx2]
    switch self.lnsItemKind {
    case LnsItemKindInt:
        return val1.(LnsInt) < val2.(LnsInt)
    case LnsItemKindReal:
        return val1.(LnsReal) < val2.(LnsReal)
    case LnsItemKindStr:
        return val1.(string) < val2.(string)
    case LnsItemKindStem:
        switch val1.(type) {
        case LnsInt:
            cval1 := val1.(LnsInt)
            switch val2.(type) {
            case LnsInt:
                return cval1 < val2.(LnsInt)
            case LnsReal:
                return LnsReal(cval1) < val2.(LnsReal)
            default:
                return true;
            }
        case LnsReal:
            cval1 := val1.(LnsReal)
            switch val2.(type) {
            case LnsInt:
                return cval1 < LnsReal(val2.(LnsInt))
            case LnsReal:
                return cval1 < val2.(LnsReal)
            default:
                return true;
            }
        case string:
            cval1 := val1.(string)
            switch val2.(type) {
            case LnsInt:
                return false;
            case LnsReal:
                return false;
            case string:
                cval2 := val2.(string)
                return cval1 < cval2;
            default:
                return true;
            }
        default:
            switch val2.(type) {
            case LnsInt:
                return false;
            case LnsReal:
                return false;
            case string:
                return false;
            default:
                return idx1 < idx2;
            }
        }
    }
    panic( "error" )
    return false
}
func (self *LnsList) Swap(idx1, idx2 int) {
    self.Items[ idx1 ], self.Items[ idx2 ] = self.Items[ idx2 ], self.Items[ idx1 ]
}


func NewLnsList( list []LnsAny ) *LnsList {
    return &LnsList{ list, LnsItemKindUnknown }
}
func (lnsList *LnsList) Insert( val LnsAny ) {
    if !Lns_IsNil( val ) {
        lnsList.Items = append( lnsList.Items, val )
    }
}
func (lnsList *LnsList) Remove( index LnsAny ) LnsAny {
    if Lns_IsNil( index ) {
        ret := lnsList.Items[ len(lnsList.Items) - 1 ]
        lnsList.Items = lnsList.Items[ : len(lnsList.Items) - 1 ]
        return ret
    } else {
        work := index.(LnsInt) - 1
        ret := lnsList.Items[ work ]
        lnsList.Items =
            append( lnsList.Items[ : work ], lnsList.Items[ work+1: ]... )
        return ret
    }
}
func (lnsList *LnsList) GetAt( index int ) LnsAny {
    return lnsList.Items[ index - 1 ]
}
func (lnsList *LnsList) Set( index int, val LnsAny ) {
    index--;
    if len( lnsList.Items ) > index {
        lnsList.Items[ index ] = val;
    } else {
        if len( lnsList.Items ) == index {
            lnsList.Items = append( lnsList.Items, val )
        } else {
            panic( fmt.Sprintf( "illegal index -- %d", index ) );
        }
    }
}
func (lnsList *LnsList) Unpack() []LnsAny {
    return lnsList.Items
}


func (LnsList *LnsList) ToLuaCode( conv *StemToLuaConv ) {
    conv.write( "{" )
    for _, val := range( LnsList.Items ) {
        conv.conv( val )
        conv.write( "," )
    }
    conv.write( "}" )
}


// ======== set ========

type LnsSet struct {
    Items map[LnsAny]bool
}

func Lns_ToSetSub(
    obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    itemParam := paramList[0]
    if val, ok := obj.(*LnsSet); ok {
        list := make([]LnsAny, len(val.Items))
        index := 0
        for key := range( val.Items ) {
            success, conved, mess :=
                itemParam.Func( key, itemParam.Nilable, itemParam.Child )
            if !success {
                return false, nil, fmt.Sprintf( "%s:%s", key, mess )
            }
            list[ index ] = conved
            index++
        }
        return true, NewLnsSet( list ), nil
    }
    return false, nil, "no set"
}

func (self *LnsSet) ToCollection() LnsAny {
    ret := NewLnsSet([]LnsAny{})
    for key := range (self.Items) {
        ret.Add( Lns_ToCollection( key ) )
    }
    return ret
}


func (self *LnsSet) CreateKeyListStem() *LnsList {
    list := make([]LnsAny, len(self.Items))
    index := 0
    for key := range self.Items {
        list[index] = key
        index++
    }    
    return NewLnsList( list )
}
func (self *LnsSet) CreateKeyListInt() *LnsList {
    list := self.CreateKeyListStem()
    list.lnsItemKind = LnsItemKindInt
    return list
}

func (self *LnsSet) CreateKeyListReal() *LnsList {
    list := self.CreateKeyListStem()
    list.lnsItemKind = LnsItemKindReal
    return list
}
func (self *LnsSet) CreateKeyListStr() *LnsList {
    list := self.CreateKeyListStem()
    list.lnsItemKind = LnsItemKindStr
    return list
}



func NewLnsSet( list []LnsAny ) *LnsSet {
    set := &LnsSet{ map[LnsAny]bool{} }
    for _, val := range( list ) {
        set.Items[ val ] = true;
    }
    return set
}

func (self *LnsSet) Add( val LnsAny ) {
    self.Items[ val ] = true;
}
func (self *LnsSet) Del( val LnsAny ) {
    delete( self.Items, val )
}
func (self *LnsSet) Has( val LnsAny ) bool {
    _, has := self.Items[ val ]
    return has
}
func (self *LnsSet) And( set *LnsSet ) *LnsSet {
    delValList := NewLnsList( []LnsAny{} )
    for val := range( self.Items ) {
        if !set.Has( val ) {
            delValList.Insert( val )
        }
    }
    for _, val := range( delValList.Items ) {
        delete( self.Items, val )
    }
    return self
}
func (self *LnsSet) Or( set *LnsSet ) *LnsSet {
    for val := range( set.Items ) {
        self.Items[ val ] = true
    }
    return self
}
func (self *LnsSet) Sub( set *LnsSet ) *LnsSet {
    delValList := NewLnsList( []LnsAny{} )
    for val := range( set.Items ) {
        if set.Has( val ) {
            delValList.Insert( val )
        }
    }
    for _, val := range( delValList.Items ) {
        delete( self.Items, val )
    }
    return self
}
func (self *LnsSet) Clone() *LnsSet {
    set := NewLnsSet( []LnsAny{} )
    for val := range( self.Items ) {
        set.Items[ val ] = true
    }
    return set
}
func (self *LnsSet) Len() LnsInt {
    return len( self.Items )
}

// ======== map ========

type LnsMap struct {
    Items map[LnsAny]LnsAny
}

func Lns_ToLnsMapSub(
    obj LnsAny, nilable bool, paramList []Lns_ToObjParam ) (bool, LnsAny, LnsAny) {
    if Lns_IsNil( obj ) {
        if nilable {
            return true, nil, nil 
        }
        return false, nil, "nil"
    }
    keyParam := paramList[0]
    itemParam := paramList[1]
    if lnsMap, ok := obj.(*LnsMap); ok {
        newMap := NewLnsMap( map[LnsAny]LnsAny{} )
        for key, val := range( lnsMap.Items ) {
            successKey, convedKey, messKey :=
                keyParam.Func( key, keyParam.Nilable, keyParam.Child )
            if !successKey {
                return false, nil, fmt.Sprintf( ".%s:%s", key,messKey)
            }
            successVal, convedVal, messVal :=
                itemParam.Func( val, itemParam.Nilable, itemParam.Child )
            if !successVal {
                return false, nil, fmt.Sprintf( ".%s:%s", val,messVal)
            }
            newMap.Items[ convedKey ] = convedVal
        }
        return true, newMap, nil
    }
    return false, nil, "no map"
}

func (self *LnsMap) ToCollection() LnsAny {
    ret := NewLnsMap( map[LnsAny]LnsAny{} )
    for key, val := range (self.Items) {
        ret.Items[ key ] = Lns_ToCollection( val )
    }
    return ret
}


func (self *LnsMap) Correct() *LnsMap {
    delete( self.Items, nil )
    list := make([]LnsAny, len(self.Items))
    index := 0
    for key, val := range self.Items {
        if Lns_IsNil( val ) {
            list[index] = key
            index++
        }
    }
    for _, key := range list[:index] {
        delete( self.Items, key )
    }
    return self
}

func (self *LnsMap) CreateKeyListStem() *LnsList {
    list := make([]LnsAny, len(self.Items))
    index := 0
    for key := range self.Items {
        list[index] = key
        index++
    }    
    return NewLnsList( list )
}
func (self *LnsMap) CreateKeyListInt() *LnsList {
    list := self.CreateKeyListStem()
    list.lnsItemKind = LnsItemKindInt
    return list
}

func (self *LnsMap) CreateKeyListReal() *LnsList {
    list := self.CreateKeyListStem()
    list.lnsItemKind = LnsItemKindReal
    return list
}
func (self *LnsMap) CreateKeyListStr() *LnsList {
    list := self.CreateKeyListStem()
    list.lnsItemKind = LnsItemKindStr
    return list
}

func NewLnsMap( arg map[LnsAny]LnsAny ) *LnsMap {
    return &LnsMap{ arg }
}


func (self *LnsMap) Set( key, val LnsAny ) {
    if Lns_IsNil( val ) {
        delete( self.Items, key );
    } else {
        self.Items[ key ] = val
    }
}

func (LnsMap *LnsMap) ToLuaCode( conv *StemToLuaConv ) {
    conv.write( "{" )
    for key, val := range( LnsMap.Items ) {
        conv.write( "[ " )
        conv.conv( key )
        conv.write( " ] =" )
        conv.conv( val )
        conv.write( "," )
    }
    conv.write( "}" )
}
