/*
    User.h
    Generated with 'sql2code' 0.0.3
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class User {

private:
    std::string id;
    std::string username;
    std::string password;
    std::string token;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string &value);
    std::string getUsername();
    void setUsername(std::string &value);
    std::string getPassword();
    void setPassword(std::string &value);
    std::string getToken();
    void setToken(std::string &value);
    int getCreated();
    void setCreated(int &value);
    int getModified();
    void setModified(int &value);
    bool isDeleted();
    void setDeleted(bool &value);
};