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

# Drogon (with Trantor) dependency:
add_subdirectory("${PROJECT_SOURCE_DIR}/../_Dependencies/Cache/Drogon/Library" "Build_Drogon")

# Drogon JWT support:
add_subdirectory("${PROJECT_SOURCE_DIR}/../_Dependencies/Cache/JWT-Drogon/Library" "Build_Drogon_JWT")

# argparse dependency:
include_directories("/usr/local/include/argparse")

file(

        GLOB APP_SRC
        "*.h" "*.cpp"
        "generated/Source/cpp/*.h" "generated/Source/cpp/*.cpp"
        "plugins/*.h" "plugins/*.cc"
)
add_executable(${VERSIONABLE_NAME} ${APP_SRC} main.cpp Utils.h Utils.cpp)

target_link_libraries(${VERSIONABLE_NAME} PRIVATE trantor)
target_link_libraries(${VERSIONABLE_NAME} PRIVATE drogon)
target_link_libraries(${VERSIONABLE_NAME} PRIVATE nlohmann_json::nlohmann_json)
target_link_libraries(${VERSIONABLE_NAME} PRIVATE cpp-jwt::cpp-jwt)

install(TARGETS ${VERSIONABLE_NAME} DESTINATION bin/${PROJECT_NAME})

install(

        FILES ${CMAKE_SOURCE_DIR}/../Configurations/default.json
        DESTINATION bin/${PROJECT_NAME} PERMISSIONS OWNER_WRITE WORLD_READ
)