<?php
class Qlimiter_Protocol_Message {
    private $_header;
    private $_meta;
    private $_type;

    public function __construct($header, $meta, $type) {
        $this->_header = $header;
        $this->_meta = $meta;
        $this->_type = $type;
    }

    public function encode() {
        $header = $this->_header->buildHeader();
        if (!isset($this->_meta['M_m']) ||  // max_val min_val
            !isset($this->_meta['M_i']) ||  // init_val
            !isset($this->_meta['M_s']) ||  // step
            !isset($this->_meta['M_t']) ||  // method
            !isset($this->_meta['M_k'])) {  // key
            throw new Exception('Invaild meta');
        }
        $metaArr = array();
        foreach ($this->_meta as $k => $v) {
            if (is_array($v)) {
                continue;
            }
            $metaArr[] = $k."\n".$v;
        }
        $metaStr = implode("\n", $metaArr);
        $buffer = $header . pack('N', strlen($metaStr)) . $metaStr;
        return $buffer;
    }

    public function decode($buffer) {

    }

    public function getHeader() {
        return $this->_header;
    }

    public function getMeta() {
        return $this->_meta;
    }
}