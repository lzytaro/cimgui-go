// Code generated by cmd/codegen from https://github.com/AllenDang/cimgui-go.
// DO NOT EDIT.

package imgui

// #include <stdlib.h>
// #include <memory.h>
// #include "extra_types.h"
// #include "cimnodes_wrapper.h"
import "C"
import "unsafe"

type EmulateThreeButtonMouse struct {
	data *C.EmulateThreeButtonMouse
}

func (self *EmulateThreeButtonMouse) handle() (result *C.EmulateThreeButtonMouse, fin func()) {
	return self.data, func() {}
}

func (self EmulateThreeButtonMouse) c() (C.EmulateThreeButtonMouse, func()) {
	result, fn := self.handle()
	return *result, fn
}

func newEmulateThreeButtonMouseFromC(cvalue *C.EmulateThreeButtonMouse) *EmulateThreeButtonMouse {
	return &EmulateThreeButtonMouse{data: cvalue}
}

type NodesContext struct {
	data *C.ImNodesContext
}

func (self *NodesContext) handle() (result *C.ImNodesContext, fin func()) {
	return self.data, func() {}
}

func newNodesContextFromC(cvalue *C.ImNodesContext) *NodesContext {
	return &NodesContext{data: cvalue}
}

type NodesEditorContext struct {
	data *C.ImNodesEditorContext
}

func (self *NodesEditorContext) handle() (result *C.ImNodesEditorContext, fin func()) {
	return self.data, func() {}
}

func newNodesEditorContextFromC(cvalue *C.ImNodesEditorContext) *NodesEditorContext {
	return &NodesEditorContext{data: cvalue}
}

type NodesIO struct {
	data *C.ImNodesIO
}

func (self *NodesIO) handle() (result *C.ImNodesIO, fin func()) {
	return self.data, func() {}
}

func (self NodesIO) c() (C.ImNodesIO, func()) {
	result, fn := self.handle()
	return *result, fn
}

func newNodesIOFromC(cvalue *C.ImNodesIO) *NodesIO {
	return &NodesIO{data: cvalue}
}

type NodesMiniMapNodeHoveringCallbackUserData struct {
	Data uintptr
}

func (self *NodesMiniMapNodeHoveringCallbackUserData) handle() (*C.ImNodesMiniMapNodeHoveringCallbackUserData, func()) {
	result, fn := self.c()
	return &result, fn
}

func (selfStruct *NodesMiniMapNodeHoveringCallbackUserData) c() (result C.ImNodesMiniMapNodeHoveringCallbackUserData, fin func()) {
	self := selfStruct.Data

	return (C.ImNodesMiniMapNodeHoveringCallbackUserData)(unsafe.Pointer(self)), func() {}
}

func newNodesMiniMapNodeHoveringCallbackUserDataFromC(cvalue *C.ImNodesMiniMapNodeHoveringCallbackUserData) *NodesMiniMapNodeHoveringCallbackUserData {
	v := (unsafe.Pointer)(*cvalue)
	return &NodesMiniMapNodeHoveringCallbackUserData{Data: uintptr(v)}
}

type NodesStyle struct {
	data *C.ImNodesStyle
}

func (self *NodesStyle) handle() (result *C.ImNodesStyle, fin func()) {
	return self.data, func() {}
}

func (self NodesStyle) c() (C.ImNodesStyle, func()) {
	result, fn := self.handle()
	return *result, fn
}

func newNodesStyleFromC(cvalue *C.ImNodesStyle) *NodesStyle {
	return &NodesStyle{data: cvalue}
}

type LinkDetachWithModifierClick struct {
	data *C.LinkDetachWithModifierClick
}

func (self *LinkDetachWithModifierClick) handle() (result *C.LinkDetachWithModifierClick, fin func()) {
	return self.data, func() {}
}

func (self LinkDetachWithModifierClick) c() (C.LinkDetachWithModifierClick, func()) {
	result, fn := self.handle()
	return *result, fn
}

func newLinkDetachWithModifierClickFromC(cvalue *C.LinkDetachWithModifierClick) *LinkDetachWithModifierClick {
	return &LinkDetachWithModifierClick{data: cvalue}
}

type MultipleSelectModifier struct {
	data *C.MultipleSelectModifier
}

func (self *MultipleSelectModifier) handle() (result *C.MultipleSelectModifier, fin func()) {
	return self.data, func() {}
}

func (self MultipleSelectModifier) c() (C.MultipleSelectModifier, func()) {
	result, fn := self.handle()
	return *result, fn
}

func newMultipleSelectModifierFromC(cvalue *C.MultipleSelectModifier) *MultipleSelectModifier {
	return &MultipleSelectModifier{data: cvalue}
}
