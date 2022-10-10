/*
    LabelTicketMapping.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "LabelTicketMapping.h"

std::string LabelTicketMapping::getId() {    
    return this->id;
}

void LabelTicketMapping::setId(std::string &value) {
    this->id = value;
}

std::string LabelTicketMapping::getLabelId() {    
    return this->labelId;
}

void LabelTicketMapping::setLabelId(std::string &value) {
    this->labelId = value;
}

std::string LabelTicketMapping::getTicketId() {    
    return this->ticketId;
}

void LabelTicketMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

int LabelTicketMapping::getCreated() {    
    return this->created;
}

void LabelTicketMapping::setCreated(int &value) {
    this->created = value;
}

int LabelTicketMapping::getModified() {    
    return this->modified;
}

void LabelTicketMapping::setModified(int &value) {
    this->modified = value;
}

bool LabelTicketMapping::isDeleted() {    
    return this->deleted;
}

void LabelTicketMapping::setDeleted(bool &value) {
    this->deleted = value;
}

