<?php
class Qlimiter_Net_Net {
    protected $_connection;
    protected $_connectionTimeout   = 0.1;
    protected $_readTimeout         = 0.2;
    protected $_readBuffer          = 256;
    protected $_writeBuffer         = 256;

    protected $_connectionStr;
    protected $_response;

    public function __construct($connectionStr) {
        $this->_connectionStr = $connectionStr;
        $this->_connection = null;
    }

    public function setConnectionTimeout($timeout) {
        $this->_connectionTimeout = $timeout;
    }

    public function setReadTimeout($timeout) {
        $this->_readTimeout = $timeout;
    }

    public function call($requestId, $params, &$ret) {
        $this->initConnection();
        if (!$this->_connection) {
            return false;
        }

        if (intval($requestId) <= 0) {
            $requestId = Qlimiter_Util_Util::genRequestId();
        }
        $this->write(Qlimiter_Protocol_Protocol::encode($requestId, $params));

        $this->_response = $this->read();
        return Qlimiter_Protocol_Protocol::isSuccess($this->_response, $ret);
    }

    protected function write($buffer) {
        $length = strlen($buffer);
        while (true) {
            $sent = fwrite($this->_connection, $buffer, $length);
            if ($sent === false || $sent === 0) {
                throw new Exception('Error to write to net'); // 这个时候可以抛异常
            }
            if ($sent < $length) {
                $buffer = substr($buffer, $sent);
                $length -= $sent;
            } else {
                return true;
            }
            usleep(5);
        }
    }

    protected function read() {
        return Qlimiter_Protocol_Protocol::decode($this->_connection);
    }

    protected function initConnection() {
        if ($this->_connection) {
            return true;
        }
        $connection = @stream_socket_client($this->_connectionStr, $code, $msg, $this->_connectionTimeout);
        if (!$connection) {
            throw new Exception('Error to connect to '.$this->_connectionStr);
        }
        $this->setStreamOpt();
        $this->_connection = $connection;
        return true;
    }

    protected function setStreamOpt() {
        if (!is_resource($this->_connection)) {
            return false;
        }
        @stream_set_timeout($this->_connection, 0, $this->_readTimeout * 1000000); // 微秒级超时，大于1s的话php内核自动转
        @stream_set_read_buffer($this->_connection, $this->_readBuffer);
        @stream_set_write_buffer($this->_connection, $this->_writeBuffer);
        return true;
    }

    public function __destruct() {
        if ($this->_connection) {
            fclose($this->_connection);
        }
    }
}
