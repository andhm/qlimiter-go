<?php
class Qlimiter_Protocol_Header {
    private $_magic;
    private $_version;
    private $_requestId;

    public function __construct($requestId, $magic, $version) {
        $this->_magic   = $magic;
        $this->_version = $version;
        $this->_requestId = $requestId;
    }

    public function buildHeader() {
        $buffer = pack('n', $this->_magic);
        $buffer .= pack('C', $this->_version);
        Qlimiter_Util_Util::split2Int($upper, $lower, $this->_requestId);
        $buffer .= pack('NN', $upper, $lower);
        return $buffer;
    }
}