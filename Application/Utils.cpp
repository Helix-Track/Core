//
// Created by milosvasic on 13.08.22.
//
#include "Utils.h"

bool Utils::isDebug() {

    return Utils::configuration.debug;
}

void Utils::setDebug(bool value) {

    Utils::configuration.debug = value;
}

bool Utils::logFull() {

    return Utils::configuration.logFull;
}

void Utils::setLogFull(bool value) {

    Utils::configuration.logFull = value;
}

void Utils::log(const std::string &tag, const std::string &message, const LogLevel &level) {

    auto time = std::chrono::duration_cast<std::chrono::milliseconds>
            (std::chrono::system_clock::now().time_since_epoch()).count();

    Utils::logger.logFull(
            VERSIONABLE_NAME,
            getVersion(),
            time,
            level,
            tag,
            message,
            std::list<std::string>()
    );
}

void Utils::v(const std::string &tag, const std::string &message) {

    auto level = LogLevel::verbose;
    log(tag, message, level);
}

void Utils::d(const std::string &tag, const std::string &message) {

    auto level = LogLevel::debug;
    log(tag, message, level);
}

void Utils::i(const std::string &tag, const std::string &message) {

    auto level = LogLevel::info;
    log(tag, message, level);
}

void Utils::w(const std::string &tag, const std::string &message) {

    auto level = LogLevel::warning;
    log(tag, message, level);
}

void Utils::e(const std::string &tag, const std::string &message) {

    auto level = LogLevel::error;
    log(tag, message, level);
}
