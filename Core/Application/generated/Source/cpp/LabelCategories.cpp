/*
    LabelCategories.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelCategories.h"

std::string LabelCategories::getId() {    
    return this->id;
}

void LabelCategories::setId(std::string value) {
    this->id = value;
}

std::string LabelCategories::getTitle() {    
    return this->title;
}

void LabelCategories::setTitle(std::string value) {
    this->title = value;
}

std::string LabelCategories::getDescription() {    
    return this->description;
}

void LabelCategories::setDescription(std::string value) {
    this->description = value;
}

int LabelCategories::getCreated() {    
    return this->created;
}

void LabelCategories::setCreated(int value) {
    this->created = value;
}

int LabelCategories::getModified() {    
    return this->modified;
}

void LabelCategories::setModified(int value) {
    this->modified = value;
}

bool LabelCategories::isDeleted() {    
    return this->deleted;
}

void LabelCategories::setDeleted(bool value) {
    this->deleted = value;
}

