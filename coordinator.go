package main

import (
    "github.com/sdwr/multi-go/socket"
    . "github.com/sdwr/multi-go/types"
)

func runCoordinator() {
    GlobalRoom.SetIncomingCallback()
    GlobalRoom.SetRegisterCallback()
    GlobalRoom.SetUnregisterCallback()
}

