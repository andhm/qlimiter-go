<?php
spl_autoload_register(function ($class) {
    if (!defined('QLIMITER_PHP_ROOT')) {
        throw new Exception("Qlimiter init Fail: should define a QLIMITER_PHP_ROOT.", 1);
    }

    $file = QLIMITER_PHP_ROOT . '/' . str_replace('_', '/', $class) . '.php';

    if (file_exists($file)) {
        require $file;
    }
});