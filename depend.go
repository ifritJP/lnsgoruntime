/*
MIT License

Copyright (c) 2018 ifritJP

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

import "os"
import "fmt"
import "runtime/pprof"

func Lns_Depend_init() {
}


/**
path の最終更新日時を取得する。

@param path ファイルパス
@return 1970/1/1 0:0:0 からの秒数。 取得失敗した場合は nil。
*/
func Depend_getFileLastModifiedTime( path string ) LnsAny {
    fileinfo, err := os.Stat( path )
    if err != nil {
        return nil
    }
    return LnsReal(fileinfo.ModTime().Unix())
}


func Depend_getLoadedMod() *Lns_luaValue {
    return Lns_getVM().GetEmptyMap()
}

func Depend_profile( validTest bool, work LnsForm, path string ) LnsAny {
    if validTest {
        prof, err := os.Create( path )
        if err != nil {
            panic( err )
        }
        if err := pprof.StartCPUProfile(prof); err != nil {
            panic( err )
        }
        defer pprof.StopCPUProfile()
    }
    
    return work( []LnsAny{} )
}

func Depend_getStackTrace() string {
   return ""
}

func Depend_searchpath( mod string, pathPattern string ) LnsAny {
    return Lns_getVM().Package_searchpath( mod, pathPattern );
}

func Depend_existFile( path string ) bool {
    if _, err := os.Stat( path ); err != nil {
        return false
    }
    return true
}

func Depend_getLuaVersion() string {
    return "5.3"
}

func Depend_canUseChannel() bool {
   return true
}
func Depend_canUseAsync() bool {
   return false
}


type Depend_UpdateVer func ( ver LnsInt )

func Depend_setup( callback Depend_UpdateVer) {
    callback( 53 );
}


var DependLuaOnLns_runLuaOnLnsFunc func(luaCode string) (LnsAny,string) = nil

func DependLuaOnLns_runLuaOnLns( luaCode string ) (LnsAny,string) {
    txt := fmt.Sprintf(`
local DependLuaOnLns = require( 'lune.base.DependLuaOnLns' )

-- 高速性最重視のため、 __luneScript に直接アクセスする
local WrapFront = {}
function WrapFront:setupFront()
   if self.readyFront then
      return
   end
   -- require( 'lune.base.front' ) 内の loadModule 処理で
   -- 再度 WrapFront:loadModule() がコールされて無限ループなるので、
   -- __luneScript を nil でクリアしておく
   __luneScript = nil
   self.readyFront = true
   local Front = require( 'lune.base.front' )
   Front.setFront()
end
function WrapFront:loadModule( mod )
   self:setupFront()
   return __luneScript:loadModule( mod )
end
function WrapFront:loadMeta( importModuleInfo, mod )
   self:setupFront()
   return __luneScript:loadMeta( importModuleInfo, mod )
end
function WrapFront:loadFromLnsTxt( importModuleInfo, name, txt )
   self:setupFront()
   return __luneScript:loadFromLnsTxt( importModuleInfo, name, txt )
end
function WrapFront:searchModule( mod )
   self:setupFront()
   return __luneScript:searchModule( mod )
end
function WrapFront:error( message )
   self:setupFront()
   return __luneScript:error( message )
end
function WrapFront.setmeta( obj )
  setmetatable( obj, { __index = WrapFront  } )
end
function WrapFront.new(  )
   local obj = {}
   WrapFront.setmeta( obj )
   return obj
end

__luneScript = WrapFront.new()

_lnsLoad = function( name, code )
   __luneScript = nil
   local frontInterface = require( 'lune.base.frontInterface' )
   local Front = require( 'lune.base.front' )
   Front.setFront()
   local importModuleInfo = frontInterface.ImportModuleInfo.new();
   return frontInterface.loadFromLnsTxt( importModuleInfo, name, code )
end

local txt=[==[
%s
]==]

return DependLuaOnLns.runLuaOnLns( txt )
`, luaCode);

    luaVM := Lns_getVM()
    loaded, err := luaVM.Load( txt, nil )
    if loaded != nil {
        ret := luaVM.RunLoadedfunc( loaded.(*Lns_luaValue), []LnsAny{} );
        return ret[ 0 ], ""
    }
    if err != nil {
        return nil, err.(string)
    }
    return nil, ""
}
func Lns_DependLuaOnLns_init() {
}
