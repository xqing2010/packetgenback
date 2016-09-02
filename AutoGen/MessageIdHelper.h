#pragma once
#include "MsgId.h"
template<typename T>
class MessageHelper {
public:
	enum  {
		Id = -1,
	};
};
template<>
class MessageHelper<CSLogin> {
public:
	enum  {
		Id = MSGID_CS_LOGIN,
	};
};
template<>
class MessageHelper<SCLoginRet> {
public:
	enum  {
		Id = MSGID_SC_LOGINRET,
	};
};
template<>
class MessageHelper<SCLoginRetsfff> {
public:
	enum  {
		Id = MSGID_SC_LOGINRETSFFF,
	};
};
