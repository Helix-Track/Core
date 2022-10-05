//
// Created by milosvasic on 13.08.22.
//

#ifndef _UTILS_H
#define _UTILS_H

#include "Logger.h"
#include "LoggerSimple.h"
#include "BuildConfig.h"
#include "VersionInfo.h"
#include "Configuration.h"

namespace Utils {

    static LoggerSimple logger;
    static Configuration configuration;

    bool isDebug();

    void setDebug(bool value);

    bool logFull();

    void setLogFull(bool value);

    void log(const std::string &tag, const std::string &message, const LogLevel &level);

    void v(const std::string &tag, const std::string &message);

    void d(const std::string &tag, const std::string &message);

    void i(const std::string &tag, const std::string &message);

    void w(const std::string &tag, const std::string &message);

    void e(const std::string &tag, const std::string &message);
}

#endif //_UTILS_H
