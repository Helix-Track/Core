/*
    LabelCategory.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelCategory.h"

std::string LabelCategory::getId() {    
    return this->id;
}

void LabelCategory::setId(std::string &value) {
    this->id = value;
}

std::string LabelCategory::getTitle() {    
    return this->title;
}

void LabelCategory::setTitle(std::string &value) {
    this->title = value;
}

std::string LabelCategory::getDescription() {    
    return this->description;
}

void LabelCategory::setDescription(std::string &value) {
    this->description = value;
}

int LabelCategory::getCreated() {    
    return this->created;
}

void LabelCategory::setCreated(int &value) {
    this->created = value;
}

int LabelCategory::getModified() {    
    return this->modified;
}

void LabelCategory::setModified(int &value) {
    this->modified = value;
}

bool LabelCategory::isDeleted() {    
    return this->deleted;
}

void LabelCategory::setDeleted(bool &value) {
    this->deleted = value;
}

