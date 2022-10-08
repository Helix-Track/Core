/*
    Comment.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Comment.h"

std::string Comment::getId() {    
    return this->id;
}

void Comment::setId(std::string value) {
    this->id = value;
}

std::string Comment::getComment() {    
    return this->comment;
}

void Comment::setComment(std::string value) {
    this->comment = value;
}

int Comment::getCreated() {    
    return this->created;
}

void Comment::setCreated(int value) {
    this->created = value;
}

int Comment::getModified() {    
    return this->modified;
}

void Comment::setModified(int value) {
    this->modified = value;
}

bool Comment::isDeleted() {    
    return this->deleted;
}

void Comment::setDeleted(bool value) {
    this->deleted = value;
}

