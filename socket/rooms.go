package socket

var GlobalRoom *Room
var QueueRoom *Room
var gameRooms map[int]*Room

func InitRooms() {
    GlobalRoom = NewRoom()
    QueueRoom = NewRoom()
    gameRooms = make(map[int]*Room)
}

func addRoom() *Room {
    room := NewRoom()
    gameRooms[room.ID] = room
    return room
}

func deleteRoom(r *Room) {
	for c, _ := range r.Clients {
		c.ChangeRoom(GlobalRoom)
	}
}
