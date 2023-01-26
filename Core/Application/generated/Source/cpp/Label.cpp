/*
    Label.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "Label.h"

std::string Label::getId() {    
    return this->id;
}

void Label::setId(std::string &value) {
    this->id = value;
}

std::string Label::getTitle() {    
    return this->title;
}

void Label::setTitle(std::string &value) {
    this->title = value;
}

std::string Label::getDescription() {    
    return this->description;
}

void Label::setDescription(std::string &value) {
    this->description = value;
}

int Label::getCreated() {    
    return this->created;
}

void Label::setCreated(int &value) {
    this->created = value;
}

int Label::getModified() {    
    return this->modified;
}

void Label::setModified(int &value) {
    this->modified = value;
}

bool Label::isDeleted() {    
    return this->deleted;
}

void Label::setDeleted(bool &value) {
    this->deleted = value;
}

