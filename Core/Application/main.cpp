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
    auto statusTag = "status";
    auto paramsTag = "parameters";

    argparse::ArgumentParser program(VERSIONABLE_NAME, getVersion());

    program.add_argument("-l", "--logFull")
            .default_value(false)
            .implicit_value(true)
            .help("Log with the full details");

    program.add_argument("-d", "--debug")
            .default_value(false)
            .implicit_value(true)
            .help("Additional information related to the parsing and code generating");

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

        if (logFull()) {

            v(paramsTag, "Full-log mode is on");
        }

        if (isDebug()) {

            w(paramsTag, "Debug mode is on");
            v(label.getTitle(), label.getDescription());
        }

        // TODO: Pass the port through the arguments
        drogon::app().addListener("0.0.0.0",8081);
        // TODO: Load config file
        //  drogon::app().loadConfigFile("../config.json");
        //  Run HTTP framework,the method will block in the internal event loop
        d(statusTag, "starting");
        drogon::app().run();

    } catch (std::logic_error &err) {

        e(errTag, err.what());
        std::exit(1);
    }
    return 0;
}