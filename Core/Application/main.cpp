#include <vector>
#include <iostream>
#include <filesystem>
#include <argparse/argparse.hpp>

#include "Utils.h"
#include "Commons.h"
#include "BuildConfig.h"
#include "VersionInfo.h"

using namespace Utils;
using namespace Commons::IO;
using namespace Commons::Strings;

int main(int argc, char *argv[]) {

    v("Hello", "World");
    return 0;
}