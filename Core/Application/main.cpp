#include <vector>
#include <iostream>
#include <filesystem>
#include <drogon/drogon.h>
#include <argparse/argparse.hpp>

#include "Utils.h"
#include "Commons.h"
#include "BuildConfig.h"

#include "VersionInfo.h"

#include "generated/Source/cpp/Label.h"

using namespace Utils;
using namespace Commons::IO;
using namespace Commons::Strings;

int main(int argc, char *argv[]) {

    std::string title = "generated code";
    std::string description = "READY";

    Label label;
    label.setTitle(title);
    label.setDescription(description);

    auto errTag = "error";
    auto startingTag = "starting";
    auto paramsTag = "parameters";
    auto noConfigurationFile = "-";

    argparse::ArgumentParser program(VERSIONABLE_NAME, getVersion());

    program.add_argument("-l", "--logFull")
            .default_value(false)
            .implicit_value(true)
            .help("Log with the full details");

    program.add_argument("-d", "--debug")
            .default_value(false)
            .implicit_value(true)
            .help("Additional information related to the parsing and code generating");

    program.add_argument("-p", "--port")
            .required()
            .scan<'i', int>()
            .help("Port to bind to");

    program.add_argument("-c", "--configurationFile")
            .default_value(std::string(noConfigurationFile))
            .help("Path to the HelixTrack core configuration file");

    program.add_argument("-g", "--logsFile")
            .default_value(std::string("HelixTrack_Logs"))
            .help("Path to the HelixTrack core logs file");

    std::string epilog("Project homepage: ");
    epilog.append(getHomepage());

    program.add_description(getDescription());
    program.add_epilog(epilog);

    try {

        program.parse_args(argc, argv);

    } catch (const std::runtime_error &err) {

        e(errTag, err.what());
        std::exit(1);
    }

    try {

        setLogFull(program["--logFull"] == true);
        setDebug(program["--debug"] == true && logFull());

        int port = program.get<int>("port");
        auto logsFile = program.get<std::string>("logsFile");
        auto configurationFile = program.get<std::string>("configurationFile");

        if (logFull()) {

            v(paramsTag, "Full-log mode is on");
        }

        if (isDebug()) {

            w(paramsTag, "Debug mode is on");
            v(label.getTitle(), label.getDescription());
        }

        v(startingTag, "Initializing");

        d(startingTag, "Logs file path: " + logsFile);

        if (configurationFile != noConfigurationFile) {

            d(startingTag, "Configuration file provided: " + configurationFile);

            /*
                Details on the configuration file:
                https://drogon.docsforge.com/master/configuration-file/
            */
            drogon::app().loadConfigFile(configurationFile);

            d(startingTag, "Ok");

        } else {

            auto logLevel = trantor::Logger::kWarn;
            if (isDebug()) {

                logLevel = trantor::Logger::kDebug;
            }

            drogon::app().addListener("0.0.0.0", port)
                    .setLogPath(logsFile)
                    .setLogLevel(logLevel);

            if (logFull()) {

                v(startingTag, "No configuration file provided");
            }

            d(startingTag, "Starting on port: " + std::to_string(port));
        }

        drogon::app().run();

    } catch (std::logic_error &err) {

        e(errTag, err.what());
        std::exit(1);

    } catch (std::runtime_error &err) {

        e(errTag, err.what());
        std::exit(1);
    }
    return 0;
}