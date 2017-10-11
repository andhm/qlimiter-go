<?php
class Qlimiter_Protocol_Protocol {
    const MAGIC     = 0xf2f2;
    const VERSION   = 0x01;

    const MSG_TYPE_REQUEST  = 0;
    const MSG_TYPE_RESPONSE = 1;

    const HEADER_BYTES      = 11;
    const META_SIZE_BYTES   = 4;

    const RES_SUCC      = 1;
    const RES_FAIL      = 0;
    const RES_ERR       = -1;

    public static function buildHeader($requestId, $version) {
        return new Qlimiter_Protocol_Header($requestId, self::MAGIC, $version);
    }

    public static function buildRequestHeader($requestId) {
        return self::buildHeader($requestId, self::VERSION);
    }

    public static function encode($requestId, $params) {
        if (!isset($params['method']) || intval($params['method']) != 0 ||
            !isset($params['initval']) || !isset($params['maxval']) ||
            intval($params['initval']) > intval($params['maxval']) ||
            !isset($params['step']) || intval($params['step']) <= 0 ||
            !isset($params['key']) || empty($params['key'])) {
            throw new Exception('Invaild meta');
        }
        $meta = array(
            'M_t'   => trim($params['method']),
            'M_k'   => trim($params['key']),
            'M_i'   => intval($params['initval']),
            'M_m'   => intval($params['maxval']),
            'M_s'   => intval($params['step']),
        );

        $header = self::buildRequestHeader($requestId);
        $msg = new Qlimiter_Protocol_Message($header, $meta, self::MSG_TYPE_REQUEST);
        return $msg->encode();
    }

    public static function decode($connection) {
        $header = fread($connection, self::HEADER_BYTES);
        if ($header === false) {
            throw new Exception('Error to read header');
        }
        if ($header === '') {
            throw new Exception('Timeout');
        }
        $header = unpack('nmagic/Cversion/NrequestIdUpper/NrequestIdLower', $header);
        $header['requestId'] = Qlimiter_Util_Util::bigInt2float($header['requestIdUpper'], $header['requestIdLower']);

        $header = self::buildHeader($header['requestId'], $header['version']);
        $metaSize = fread($connection, self::META_SIZE_BYTES);
        if ($metaSize === false) {
            throw new Exception('Error to read meta-size');
        }
        $metaSize = unpack('NmetaSize', $metaSize);
        $meta = array();
        if ($metaSize['metaSize'] > 0) {
            $metaBuf = fread($connection, $metaSize['metaSize']);
            if ($metaBuf === false) {
                throw new Exception('Error to read meta');
            }
            $metaArr = explode("\n", unpack('A*meta', $metaBuf)['meta']);
            for ($i = 0; $i < count($metaArr); $i++) {
                $meta[$metaArr[$i]] = $metaArr[++$i];
            }
        }

        return new Qlimiter_Protocol_Message($header, $meta, self::MSG_TYPE_RESPONSE);
    }

    public static function isSuccess($response, &$retCurrVal) {
        $meta = $response->getMeta();
        if (!isset($meta['M_r']) || !isset($meta['M_e'])) {
            throw new Exception('Error or Timeout');
        } else if ($meta['M_r'] == self::RES_ERR) {
            throw new Exception($meta['M_e']);
        }
        $retCurrVal = intval($meta['M_c']);
        return intval($meta['M_r']) == self::RES_SUCC;
    }
}