/*
    Report.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Report.h"

std::string Report::getId() {    
    return this->id;
}

void Report::setId(std::string &value) {
    this->id = value;
}

int Report::getCreated() {    
    return this->created;
}

void Report::setCreated(int &value) {
    this->created = value;
}

int Report::getModified() {    
    return this->modified;
}

void Report::setModified(int &value) {
    this->modified = value;
}

std::string Report::getTitle() {    
    return this->title;
}

void Report::setTitle(std::string &value) {
    this->title = value;
}

std::string Report::getDescription() {    
    return this->description;
}

void Report::setDescription(std::string &value) {
    this->description = value;
}

bool Report::isDeleted() {    
    return this->deleted;
}

void Report::setDeleted(bool &value) {
    this->deleted = value;
}

