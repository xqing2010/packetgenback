package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var inputDir *string
var outputDir *string
var exportFlag *string

func init() {
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

func isFileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func clearDir(path string) {
	files, err := ioutil.ReadDir(path)
	if nil != err {
		fmt.Println(err)
		return
	}

	for _, f := range files {
		err := os.Remove(path + "/" + f.Name())
		if nil != err {
			fmt.Println(err)
		}
	}
}

func genAutoGenPrefix() error {
	filename := *outputDir + "/" + "MsgID.h"
	var buff bytes.Buffer
	file, err := os.Create(filename)
	if nil != err {
		return err
	}
	buff.WriteString("#pragma once\n\n")
	buff.WriteString("enum E_MSGID {\n")
	buff.WriteString("\tMSGID_INVALID = -1,\n")
	buff.WriteString("\t// begin auto gen Id-----\n")
	file.WriteString(buff.String())
	file.Close()

	filename = *outputDir + "/" + "MessageIdHelper.h"
	file, err = os.Create(filename)
	if nil != err {
		return err
	}
	buff.Reset()
	buff.WriteString("#pragma once\n")
	buff.WriteString("#include \"PBMessage.pb.h\"\n")
	buff.WriteString("#include \"AutoGen/PBMsgId.pb.h\"\n\n")
	buff.WriteString("template<typename T>\n")
	buff.WriteString("class MessageHelper {\n")
	buff.WriteString("public:\n")
	buff.WriteString("\tenum  {\n")
	buff.WriteString("\t\tId = -1,\n")
	buff.WriteString("\t};\n")
	buff.WriteString("};\n\n")
	file.WriteString(buff.String())
	file.Close()

	return nil
}

func genAutoGenSuffix() error {
	filename := *outputDir + "/" + "MsgID.h"
	file, err := os.OpenFile(filename, os.O_APPEND, os.ModePerm)
	if nil != err {
		return err
	}
	var buff bytes.Buffer
	buff.WriteString("\t// end-----\n")
	buff.WriteString("\tMSGID_MAX_SIZE,\n")
	buff.WriteString("};\n")
	file.WriteString(buff.String())

	return nil
}

func AutoConvert() error {
	/*	dirName := *inputDir
		//dirName = "F:/tmp"
		files, err := ioutil.ReadDir(dirName)
		if nil != err {
			fmt.Println(err)
		}
		var buff bytes.Buffer

	*/
	var buff bytes.Buffer

	pathname := *inputDir

	files, err := ioutil.ReadDir(pathname)
	if nil != err {
		fmt.Println(err)
	}

	if false == isDirExists(*outputDir) {
		err := os.Mkdir(*outputDir, os.ModePerm)
		if nil != err {
			fmt.Println(err)
			return err
		}
	} else {
		clearDir(*outputDir)
	}

	genAutoGenPrefix()
	for _, finfo := range files {
		pathname := pathname + "/" + finfo.Name()
		datas, err := ioutil.ReadFile(pathname)
		if nil != err {
			fmt.Println(err)
		}
		buff.Reset()
		buff.Write(datas)
		str := buff.String()
		fileGen(str, finfo.Name())
	}
	genAutoGenSuffix()
	return err
}

func fileGen(content, filename string) error {
	lines := strings.Split(content, "\r\n")
	messages := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		index := strings.Index(line, "message")
		if 0 == index {
			splits := strings.Split(line, " ")
			index = strings.Index(splits[1], "{")
			var messageName string
			if -1 != index {
				messageName = splits[1][0:index]
			} else {
				messageName = splits[1][0:]
			}
			messages = append(messages, messageName)
		}
	}

	genMessageIdHelper(messages)
	//genMessageId(messages)
	genRegisterPacket(messages)
	genHandlerDeclare(messages)
	genMsgIdMap(messages)

	return nil
}

func getMessageId(messageName string) string {
	var messageId string
	prefix := messageName[0:2]
	name := messageName[2:]
	messageId = "MSGID_" + strings.ToUpper(prefix) + "_" + strings.ToUpper(name)
	return messageId
}

func genMessageId(messages []string) error {
	filename := *outputDir + "/" + "MsgID.h"

	var buff bytes.Buffer

	file, err := os.OpenFile(filename, os.O_APPEND, os.ModePerm)
	if nil != err {
		return err
	}

	defer file.Close()

	for _, message := range messages {
		buff.WriteString("\t" + getMessageId(message) + ",\n")
	}

	file.WriteString(buff.String())
	return nil
}

func genMessageIdHelper(messages []string) error {
	filename := *outputDir + "/" + "MessageIdHelper.h"
	file, err := os.OpenFile(filename, os.O_APPEND, os.ModePerm)
	if nil != err {
		return err
	}
	defer file.Close()

	var buff bytes.Buffer

	for _, messageName := range messages {
		prefix := messageName[0:2]
		cut := strings.ToUpper(prefix)
		if cut != "PB" {
			buff.WriteString("template<>\n")
			buff.WriteString("class MessageHelper<" + messageName + "> {\n")
			buff.WriteString("public:\n")
			buff.WriteString("\tenum  {\n")
			buff.WriteString("\t\tId = " + "MsgId::MsgId_" + messageName + ",\n")
			buff.WriteString("\t};\n")
			buff.WriteString("};\n\n")
		}
	}
	file.WriteString(buff.String())
	return nil
}

func isExport(message string) bool {
	prefix := message[0:2]
	cut := strings.ToUpper(prefix)
	if *exportFlag == "gs" {
		if cut != "CS" && cut != "DG" {
			return false
		}
	} else if *exportFlag == "ds" {
		if cut != "GD" {
			return false
		}
	} else if *exportFlag == "sc" {
		if cut != "SC" {
			return false
		}
	}

	return true
}

func genRegisterPacket(messages []string) error {
	filename := *outputDir + "/" + "register.inl"
	file, err := os.OpenFile(filename, os.O_APPEND, os.ModePerm)
	if nil != err {
		file, err = os.Create(filename)
		if nil != err {
			return err
		}
	}
	defer file.Close()

	var buff bytes.Buffer

	for _, message := range messages {
		if isExport(message) == false {
			continue
		}
		file.WriteString("\tREGISTER_PACKETFACTORY_WITH_HANDLER(new PacketFactory_T<PBPacket<" + message + "> >, (pPacketHandlerFunc)MSG_HANDLER::on" + message + ");\n")
	}

	file.WriteString(buff.String())

	return nil
}

func genHandlerDeclare(messages []string) error {
	filename := *outputDir + "/" + "handlerdeclare.inl"
	file, err := os.OpenFile(filename, os.O_APPEND, os.ModePerm)
	if nil != err {
		file, err = os.Create(filename)
		if nil != err {
			return err
		}
	}
	defer file.Close()

	var buff bytes.Buffer

	for _, message := range messages {
		if isExport(message) == false {
			continue
		}
		file.WriteString("\textern MSG_HANDLER_RETURN on" + message + "(IMsgHandler* pMsgHandler, Packet* pMsg);\n")
	}

	file.WriteString(buff.String())

	return nil
}

func genMsgIdMap(messages []string) error {
	filename := *outputDir + "/" + "msgidmap.lua"
	file, err := os.Create(filename)
	if nil != err {
		return err
	}
	defer file.Close()

	var buff bytes.Buffer
	file.WriteString("local messagetable = {\n")
	num := len(messages)
	for id, message := range messages {

		file.WriteString("[" + strconv.Itoa(id) + "] = \"" + message + "\"")
		if id != num-1 {
			file.WriteString(",\n")
		} else {
			file.WriteString("\n")
		}
		id++
	}
	file.WriteString("}\n")
	file.WriteString("return messagetable\n")

	file.WriteString(buff.String())

	return nil
}

func genPBMsgId(messages []string) error {
	filename := *outputDir + "/" + "PBMsgId.proto"
	file, err := os.Create(filename)
	if nil != err {
		return err
	}
	defer file.Close()
	var buff bytes.Buffer
	buff.WriteString("syntax = \"proto3\";\n")
	buff.WriteString("option optimize_for = LITE_RUNTIME;\n")
	buff.WriteString("enum MsgId {\n")
	for id, message := range messages {
		buff.WriteString("\tMsgId_" + message + " = " + strconv.Itoa(id) + ",\n")
	}
	buff.WriteString("}\n")
	file.WriteString(buff.String())
	return nil
}

func main() {
	inputDir = flag.String("idir", "Proto", "idir path: input dir!")
	outputDir = flag.String("odir", "AutoGen", "odir path: output dir!")
	exportFlag = flag.String("export", "gs", "export flag: flag value gs for gameserver, ds for dataserver, sc for client")

	flag.Parse()
	/*
	*inputDir = "F:\\project_modify\\server\\ClothesChange_Common\\proto\\input"
	*outputDir = "F:\\project_modify\\server\\ClothesChange_Common\\proto\\out"
	 */
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	if false == strings.Contains(*inputDir, dir) {
		*inputDir = dir + *inputDir
	}

	if false == strings.Contains(*outputDir, dir) {
		*outputDir = dir + *outputDir
	}

	for _, arg := range flag.Args() {
		fmt.Println(arg)
	}
	AutoConvert()
}
