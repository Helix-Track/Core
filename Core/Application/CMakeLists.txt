cmake_minimum_required(VERSION 3.20)

include(${CMAKE_CURRENT_SOURCE_DIR}/../Versionable/CMake/CMakeLists.txt)

project(${PROJECT_FULL_NAME})

include_directories("${CMAKE_CURRENT_SOURCE_DIR}")

# Logger dependency:
include_directories("/usr/local/include/Logger-1.0.0")
link_directories("/usr/local/lib/Logger-1.0.0")
link_libraries(Logger-1.0.0)

# Commons dependency:
include_directories("/usr/local/include/Commons-1.0.0")
link_directories("/usr/local/lib/Commons-1.0.0")
link_libraries(Commons-1.0.0)

# Tester dependency:
include_directories("/usr/local/include/Tester-1.0.0")
link_directories("/usr/local/lib/Tester-1.0.0")
link_libraries(Tester-1.0.0)

# argparse dependency:
include_directories("/usr/local/include/argparse")

file(

        GLOB APP_SRC
        "*.h" "*.cpp"
        "generated/Source/cpp/*.h" "generated/Source/cpp/*.cpp"
)
add_executable(${VERSIONABLE_NAME} ${APP_SRC} main.cpp Utils.h Utils.cpp)

install(TARGETS ${VERSIONABLE_NAME} DESTINATION bin/${PROJECT_NAME})