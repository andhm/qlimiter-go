<?php
class Qlimiter_Client {
    private $_clientHander;
    private $_host;
    private $_port;

    public function __construct($host, $port) {
        $this->_clientHander = new Qlimiter_Net_Net('tcp://'.$host.':'.$port);
        $this->_host = $host;
        $this->_port = $port;
    }

    public function setConnectionTimeout($timout) {
        $this->_clientHander->setConnectionTimeout($timeout);
    }

    public function setReadTimeout($timeout) {
        $this->_clientHander->setReadTimeout($timeout);
    }

    public function getClientHander() {
        return $this->_clientHander;
    }

    public function getHost() {
        return $this->_host;
    }

    public function getPort() {
        return $this->_port;
    }

    public function limit($key, $max, &$ret) {
        $params = array(
            'method'    => 0,
            'key'       => trim($key),
            'maxval'    => intval($max),
            'initval'   => 0,
            'step'      => 1,
        );

        return $this->_clientHander->call(0, $params, $ret);
    }
}