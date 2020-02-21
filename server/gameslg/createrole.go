package gameslg

import (
	"encoding/json"
	"github.com/llr104/LiFrame/core/liFace"
	"github.com/llr104/LiFrame/core/liNet"
	"github.com/llr104/LiFrame/server/db/slgdb"
	"github.com/llr104/LiFrame/server/gameslg/slgproto"
	"github.com/llr104/LiFrame/utils"
)

var CreateRole createRole

func init() {
	CreateRole = createRole{}
}
type createRole struct {
	liNet.BaseRouter
}

func (s *createRole) NameSpace() string {
	return "birth"
}

func (s *createRole) QryRoleReq(req liFace.IRequest)  {
	ackInfo := slgproto.QryRoleAck{}
	if p, err := req.GetConnection().GetProperty("roleId"); err == nil {

		roleId := p.(uint32)
		role := playerMgr.getRole(roleId)
		if role == nil{
			ackInfo.Code = slgproto.Code_Role_Not_Found
		}else{
			ackInfo.Code = slgproto.Code_SLG_Success
			ackInfo.Role = *role
		}
		data, _ := json.Marshal(ackInfo)
		req.GetConnection().SendMsg(slgproto.BirthQryRoleAck, data)

	}else{

		if p, err := req.GetConnection().GetProperty("userId"); err != nil{
			ackInfo.Code = slgproto.Code_Not_Auth
		}else{
			userId := p.(uint32)
			r := slgdb.NewDefaultRole()
			r.UserId = userId
			if err := slgdb.FindRoleByUserId(&r); err == nil{
				playerMgr.createPlayer(&r)
				ackInfo.Role = r
				ackInfo.Code = slgproto.Code_SLG_Success
				req.GetConnection().SetProperty("roleId", r.RoleId)
			}else{
				ackInfo.Code = slgproto.Code_Role_Not_Found
			}
		}

		data, _ := json.Marshal(ackInfo)
		req.GetConnection().SendMsg(slgproto.BirthQryRoleAck, data)
	}
}

func (s *createRole) NewRoleReq(req liFace.IRequest) {
	reqInfo := slgproto.NewRoleReq{}
	ackInfo := slgproto.NewRoleAck{}
	json.Unmarshal(req.GetData(), &reqInfo)
	p, err := req.GetConnection().GetProperty("userId")

	if err != nil{
		ackInfo.Code = slgproto.Code_Not_Auth
	}else{
		//创建角色
		userId := p.(uint32)
		r := slgdb.NewDefaultRole()
		r.Name = reqInfo.Name
		r.UserId = userId
		r.Nation = reqInfo.Nation
		if err := slgdb.FindRoleByName(&r); err == nil{
			ackInfo.Code = slgproto.Code_Role_Exit
		}else{

			if id, err := slgdb.InsertRoleToDB(&r); err == nil {
				ackInfo.Role = r
				ackInfo.Code = slgproto.Code_SLG_Success
				req.GetConnection().SetProperty("roleId", r.RoleId)
				//创建好角色直接开放所有的建筑
				arr1 := slgdb.NewRoleAllDwellings(uint32(id))
				slgdb.InsertDwellingsToDB(arr1)

				arr2 := slgdb.NewRoleAllBarracks(uint32(id))
				slgdb.InsertBarracksToDB(arr2)

				arr3 := slgdb.NewRoleAllBLumbers(uint32(id))
				slgdb.InsertLumbersToDB(arr3)

				arr4 := slgdb.NewRoleAllBFarmlands(uint32(id))
				slgdb.InsertFarmlandsToDB(arr4)

				arr5 := slgdb.NewRoleAllMines(uint32(id))
				slgdb.InsertMinesToDB(arr5)

				playerMgr.addPlayer(&r, arr2, arr1, arr4, arr3, arr5)

				utils.Log.Info("new role:%d", id)
			}else {
				ackInfo.Code = slgproto.Code_DB_Error
				utils.Log.Info("new role error: %s", err.Error())
			}
		}
	}

	data, _ := json.Marshal(ackInfo)
	req.GetConnection().SendMsg(slgproto.BirthNewRoleAck, data)
}
