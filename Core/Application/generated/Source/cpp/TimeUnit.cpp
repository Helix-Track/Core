/*
    TimeUnit.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "TimeUnit.h"

std::string TimeUnit::getId() {    
    return this->id;
}

void TimeUnit::setId(std::string &value) {
    this->id = value;
}

std::string TimeUnit::getTitle() {    
    return this->title;
}

void TimeUnit::setTitle(std::string &value) {
    this->title = value;
}

std::string TimeUnit::getDescription() {    
    return this->description;
}

void TimeUnit::setDescription(std::string &value) {
    this->description = value;
}

int TimeUnit::getCreated() {    
    return this->created;
}

void TimeUnit::setCreated(int &value) {
    this->created = value;
}

int TimeUnit::getModified() {    
    return this->modified;
}

void TimeUnit::setModified(int &value) {
    this->modified = value;
}

bool TimeUnit::isDeleted() {    
    return this->deleted;
}

void TimeUnit::setDeleted(bool &value) {
    this->deleted = value;
}

