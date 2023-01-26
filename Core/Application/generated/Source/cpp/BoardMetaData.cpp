/*
    BoardMetaData.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "BoardMetaData.h"

std::string BoardMetaData::getId() {    
    return this->id;
}

void BoardMetaData::setId(std::string &value) {
    this->id = value;
}

std::string BoardMetaData::getBoardId() {    
    return this->boardId;
}

void BoardMetaData::setBoardId(std::string &value) {
    this->boardId = value;
}

std::string BoardMetaData::getProperty() {    
    return this->property;
}

void BoardMetaData::setProperty(std::string &value) {
    this->property = value;
}

std::string BoardMetaData::getValue() {    
    return this->value;
}

void BoardMetaData::setValue(std::string &value) {
    this->value = value;
}

int BoardMetaData::getCreated() {    
    return this->created;
}

void BoardMetaData::setCreated(int &value) {
    this->created = value;
}

int BoardMetaData::getModified() {    
    return this->modified;
}

void BoardMetaData::setModified(int &value) {
    this->modified = value;
}

