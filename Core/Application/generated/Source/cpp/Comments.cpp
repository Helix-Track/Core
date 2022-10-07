/*
    Comments.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Comments.h"

std::string Comments::getId() {    
    return this->id;
}

void Comments::setId(std::string value) {
    this->id = value;
}

std::string Comments::getComment() {    
    return this->comment;
}

void Comments::setComment(std::string value) {
    this->comment = value;
}

int Comments::getCreated() {    
    return this->created;
}

void Comments::setCreated(int value) {
    this->created = value;
}

int Comments::getModified() {    
    return this->modified;
}

void Comments::setModified(int value) {
    this->modified = value;
}

bool Comments::isDeleted() {    
    return this->deleted;
}

void Comments::setDeleted(bool value) {
    this->deleted = value;
}

