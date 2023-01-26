/*
    ComponentTicketMapping.cpp
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "ComponentTicketMapping.h"

std::string ComponentTicketMapping::getId() {    
    return this->id;
}

void ComponentTicketMapping::setId(std::string &value) {
    this->id = value;
}

std::string ComponentTicketMapping::getComponentId() {    
    return this->componentId;
}

void ComponentTicketMapping::setComponentId(std::string &value) {
    this->componentId = value;
}

std::string ComponentTicketMapping::getTicketId() {    
    return this->ticketId;
}

void ComponentTicketMapping::setTicketId(std::string &value) {
    this->ticketId = value;
}

int ComponentTicketMapping::getCreated() {    
    return this->created;
}

void ComponentTicketMapping::setCreated(int &value) {
    this->created = value;
}

int ComponentTicketMapping::getModified() {    
    return this->modified;
}

void ComponentTicketMapping::setModified(int &value) {
    this->modified = value;
}

bool ComponentTicketMapping::isDeleted() {    
    return this->deleted;
}

void ComponentTicketMapping::setDeleted(bool &value) {
    this->deleted = value;
}

