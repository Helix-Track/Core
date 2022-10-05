/*
    Boards.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Boards.h"

std::string Boards::getId() {    
    return this->id;
}

void Boards::setId(std::string value) {
    this->id = value;
}

std::string Boards::getTitle() {    
    return this->title;
}

void Boards::setTitle(std::string value) {
    this->title = value;
}

std::string Boards::getDescription() {    
    return this->description;
}

void Boards::setDescription(std::string value) {
    this->description = value;
}

int Boards::getCreated() {    
    return this->created;
}

void Boards::setCreated(int value) {
    this->created = value;
}

int Boards::getModified() {    
    return this->modified;
}

void Boards::setModified(int value) {
    this->modified = value;
}

bool Boards::isDeleted() {    
    return this->deleted;
}

void Boards::setDeleted(bool value) {
    this->deleted = value;
}

