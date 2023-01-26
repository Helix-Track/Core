/*
    Team.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Team.h"

std::string Team::getId() {    
    return this->id;
}

void Team::setId(std::string &value) {
    this->id = value;
}

std::string Team::getTitle() {    
    return this->title;
}

void Team::setTitle(std::string &value) {
    this->title = value;
}

std::string Team::getDescription() {    
    return this->description;
}

void Team::setDescription(std::string &value) {
    this->description = value;
}

int Team::getCreated() {    
    return this->created;
}

void Team::setCreated(int &value) {
    this->created = value;
}

int Team::getModified() {    
    return this->modified;
}

void Team::setModified(int &value) {
    this->modified = value;
}

bool Team::isDeleted() {    
    return this->deleted;
}

void Team::setDeleted(bool &value) {
    this->deleted = value;
}

