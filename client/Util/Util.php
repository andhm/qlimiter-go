<?php
class Qlimiter_Util_Util {
    const BIGINT_DIVIDER = 0xffffffff;

    public static function genRequestId() {
        $time = explode(" ", microtime());
        $rnd = rand(0, 999);
        $request_id = sprintf("%d%06d%03d", $time[1], (int) ($time[0]*1000000), $rnd);
        return $request_id;
    }

    public static function split2Int(&$upper, &$lower, $value) {
        $lower = intval($value % (self::BIGINT_DIVIDER + 1));
        $upper = intval(($value - $lower) / (self::BIGINT_DIVIDER + 1));
    }

    public static function bigInt2float($upper, $lower) {
        return $upper * (self::BIGINT_DIVIDER + 1) + $lower;
    }
}