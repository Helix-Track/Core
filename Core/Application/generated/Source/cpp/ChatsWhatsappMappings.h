/*
    ChatsWhatsappMappings.h
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class ChatsWhatsappMappings {

private:
    std::string id;
    std::string chatId;
    std::string property;
    std::string value;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getChatId();
    void setChatId(std::string value);
    std::string getProperty();
    void setProperty(std::string value);
    std::string getValue();
    void setValue(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};