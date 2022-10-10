/*
    SystemInfo.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "SystemInfo.h"

int SystemInfo::getId() {    
    return this->id;
}

void SystemInfo::setId(int &value) {
    this->id = value;
}

std::string SystemInfo::getDescription() {    
    return this->description;
}

void SystemInfo::setDescription(std::string &value) {
    this->description = value;
}

int SystemInfo::getCreated() {    
    return this->created;
}

void SystemInfo::setCreated(int &value) {
    this->created = value;
}

