/*
    LabelTeamMapping.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelTeamMapping.h"

std::string LabelTeamMapping::getId() {    
    return this->id;
}

void LabelTeamMapping::setId(std::string &value) {
    this->id = value;
}

std::string LabelTeamMapping::getLabelId() {    
    return this->labelId;
}

void LabelTeamMapping::setLabelId(std::string &value) {
    this->labelId = value;
}

std::string LabelTeamMapping::getTeamId() {    
    return this->teamId;
}

void LabelTeamMapping::setTeamId(std::string &value) {
    this->teamId = value;
}

int LabelTeamMapping::getCreated() {    
    return this->created;
}

void LabelTeamMapping::setCreated(int &value) {
    this->created = value;
}

int LabelTeamMapping::getModified() {    
    return this->modified;
}

void LabelTeamMapping::setModified(int &value) {
    this->modified = value;
}

bool LabelTeamMapping::isDeleted() {    
    return this->deleted;
}

void LabelTeamMapping::setDeleted(bool &value) {
    this->deleted = value;
}

