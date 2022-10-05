#include <vector>
#include <iostream>
#include <filesystem>
#include <argparse/argparse.hpp>

#include "Utils.h"
#include "Commons.h"
#include "BuildConfig.h"
#include "VersionInfo.h"

#include "generated/Source/cpp/Projects.h"

using namespace Utils;
using namespace Commons::IO;
using namespace Commons::Strings;

int main(int argc, char *argv[]) {

    Projects project;
    project.setTitle("Hello");
    project.setIdentifier("World");

    v(project.getTitle(), project.getIdentifier());
    return 0;
}