#!/usr/bin/python
# Copyright 2024-2025 NetCracker Technology Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import binascii
import logging
import struct
import time

SESSIONCLOSE = -11
SESSIONCREATE = -10
ERROR = -1
NOTIFICATION = 0
CREATE = 1
DELETE = 2
EXISTS = 3
GETDATA = 4
SETDATA = 5
GETACL = 6
SETACL = 7
GETCHILDREN = 8
SYNC = 9
PING = 11
GETCHILDREN2 = 12
CHECK = 13
MULTI = 14
AUTH = 100
SETWATCHES = 101
SASL = 102

opcodes = {
    SESSIONCLOSE: 'sessionclose',
    SESSIONCREATE: 'sessioncreate',
    ERROR: 'error',
    NOTIFICATION: 'notification',
    CREATE: 'create',
    DELETE: 'delete',
    EXISTS: 'exists',
    GETDATA: 'getdata',
    SETDATA: 'setdata',
    GETACL: 'getacl',
    SETACL: 'setacl',
    GETCHILDREN: 'getchildren',
    SYNC: 'sync',
    PING: 'ping',
    GETCHILDREN2: 'getchildren2',
    CHECK: 'check',
    MULTI: 'multi',
    AUTH: 'auth',
    SETWATCHES: 'setwatches',
    SASL: 'sasl',
}

END_OF_STREAM = b'\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00'


# end of stream
class EOS(Exception):
    pass


class UnknownType(Exception):
    def __init__(self, operation_type):
        self.type = operation_type

    def __str__(self):
        return f'Unknown type {self.type}'


class LogFileHeader(object):

    def __init__(self, stream):
        magic_struct = struct.Struct('4s')
        self.MAGIC = int(binascii.hexlify(magic_struct.pack(b'ZKLG')), 16)
        data_struct = struct.Struct('>i i q')
        self.data_bytes = stream.read(data_struct.size)
        self.magic, self.version, self.dbid = data_struct.unpack(self.data_bytes)

    def is_valid(self):
        return self.magic == self.MAGIC


class TransactionData(object):
    def __init__(self, data):
        self.initial = data
        self.body = data
        self.rest = data

    def read(self, count):
        part = self.rest[:count]
        self.rest = self.rest[count:]
        logging.debug(f'Read by TransactionData is {part}.')
        return part

    def __str__(self):
        return f'Initial is {self.initial}, rest is {self.rest}'


class Txn(object):

    def __init__(self, stream):
        s = struct.Struct('>q i')
        txn_head = stream.read(s.size)
        self.crc, self.txn_len = s.unpack(txn_head)

        if not self.txn_len:
            raise EOS()

        transaction_data = TransactionData(stream.read(self.txn_len))

        self.header = h = TxnHeader(transaction_data)
        if h.type == CREATE:
            self.entry = TxnCreate(transaction_data)
        elif h.type == DELETE:
            self.entry = TxnDelete(transaction_data)
        elif h.type == SETDATA:
            self.entry = TxnSetData(transaction_data)
        elif h.type == SETACL:
            self.entry = TxnSetAcl(transaction_data)
        elif h.type == SESSIONCREATE:
            self.entry = TxnSessionCreate(transaction_data)
        elif h.type == SESSIONCLOSE:
            self.entry = TxnSessionClose(transaction_data)
        elif h.type == ERROR:
            self.entry = TxnError(transaction_data)
        elif h.type == MULTI:
            self.entry = None
        else:
            raise UnknownType(h.type)

        eor = stream.read(1)

        self.transaction_bytes = txn_head + transaction_data.initial + eor

        if h.type == SESSIONCREATE or h.type == SESSIONCLOSE or h.type == CREATE and self.entry.ephemeral == 1:
            self.transaction_bytes = b''

    def __str__(self):
        return f'{self.header} -- {self.entry}' if self.entry else f'{self.header} -- Unrecognized operation'


class TxnHeader(object):

    def __init__(self, record):
        s = struct.Struct('>Q I Q Q i')
        logging.debug('Header bytes are %s.', record)
        self.client_id, self.cxid, self.zxid, self.time, self.type = s.unpack(record.read(s.size))
        record.body = record.rest

    @staticmethod
    def op2type(operation_type):
        return opcodes[operation_type]

    def __str__(self):
        return "%s (%3dms) sessionid 0x%x zxid 0x%x cxid 0x%x %s" % (
            time.ctime(self.time / 1000), self.time % 1000, self.client_id, self.zxid,
            self.cxid, self.op2type(self.type))


class TxnEntry(object):

    def read_string(self, record):
        length = self.read_int(record)
        s = struct.Struct(str(length) + 's')
        string = s.unpack(record.read(s.size))
        return string

    def read_data(self, record):
        length = self.read_int(record)
        return record.read(length)

    def read_acls(self, record):
        logging.debug('Read acl')
        count = self.read_int(record)
        return [self.read_acl(record) for _ in range(count)]

    @staticmethod
    def read_acl(record):
        return Acl(record)

    @staticmethod
    def read_int(record):
        s = struct.Struct('>i')
        integer, = s.unpack(record.read(s.size))
        return integer

    @staticmethod
    def read_bool(record):
        s = struct.Struct('B')
        boolean = s.unpack(record.read(s.size))
        return boolean == 0


class TxnCreate(TxnEntry):

    def __init__(self, record):
        try:
            self.path = self.read_string(record)
            self.data = ''
            self.acls = self.read_acls(record)
            self.ephemeral = self.read_bool(record)
        except struct.error:
            logging.debug('Creation record contains "data", so it should be taken into account when parsing.')
            record.rest = record.body
            self.path = self.read_string(record)
            self.data = self.read_data(record)
            self.acls = self.read_acls(record)
            self.ephemeral = self.read_bool(record)

    def __str__(self):
        return f"Create path {self.path} data '{self.data}' acls - {self.acls} ephemeral {self.ephemeral}"


class TxnDelete(TxnEntry):

    def __init__(self, record):
        self.path = self.read_string(record)

    def __str__(self):
        return "Delete path %s" % self.path


class TxnSetData(TxnEntry):

    def __init__(self, record):
        self.path = self.read_string(record)
        self.data = self.read_data(record)
        self.version = self.read_int(record)

    def __str__(self):
        return f"SetData path {self.path} data '{self.data}' version {self.version}"


class TxnSetAcl(TxnEntry):

    def __init__(self, record):
        self.path = self.read_string(record)
        self.acls = self.read_acls(record)
        self.version = self.read_int(record)

    def __str__(self):
        return f"SetAcl path {self.path} acls - {self.acls} version {self.version}"


class TxnSessionCreate(TxnEntry):

    def __init__(self, record):
        self.timeout = self.read_int(record)

    def __str__(self):
        return "SessionCreate timeout %ims" % self.timeout


class TxnSessionClose(TxnEntry):

    def __init__(self, record):
        pass

    def __str__(self):
        return "SessionClose"


class Acl(TxnEntry):
    def __init__(self, record):
        self.perms = self.read_int(record)
        self.scheme = self.read_string(record)
        self.id = self.read_string(record)

    def __str__(self):
        return "Acl %s %s %x" % (self.scheme, self.id, self.perms)


class TxnError(TxnEntry):
    Ok = 0
    SystemError = -1
    RuntimeInconsistency = -2
    DataInconsistency = -3
    ConnectionLoss = -4
    MarshallingError = -5
    Unimplemented = -6
    OperationTimeout = -7
    BadArguments = -8
    APIError = -100
    NoNode = -101
    NoAuth = -102
    BadVersion = -103
    NoChildrenForEphemerals = -108
    NodeExists = -110
    NotEmpty = -111
    SessionExpired = -112
    InvalidCallback = -113
    InvalidACL = -114
    AuthFailed = -115
    SessionMoved = -118

    errorcodes = {
        Ok: 'ok',
        SystemError: 'systemerror',
        RuntimeInconsistency: 'runtimeinconsistency',
        DataInconsistency: 'datainconsistency',
        ConnectionLoss: 'connectionloss',
        MarshallingError: 'marshallingerror',
        Unimplemented: 'unimplemented',
        OperationTimeout: 'operationtimeout',
        BadArguments: 'badarguments',
        APIError: 'apierror',
        NoNode: 'nonode',
        NoAuth: 'noauth',
        BadVersion: 'badversion',
        NoChildrenForEphemerals: 'nochildrenforephemerals',
        NodeExists: 'nodeexists',
        NotEmpty: 'notempty',
        SessionExpired: 'sessionexpired',
        InvalidCallback: 'invalidcallback',
        InvalidACL: 'invalidacl',
        AuthFailed: 'authfailed',
        SessionMoved: 'sessionmoved',
    }

    def __init__(self, record):
        self.err = self.read_int(record)

    def __str__(self):
        return f'Error {self.errorcodes[self.err]}'
