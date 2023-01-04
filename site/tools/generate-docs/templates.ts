import fs from 'fs';
import Handlebars from 'handlebars';
import path from 'path';
import { camelCase, pascalCase, snakeCase } from './utils';

const TEMPLATES_DIR = path.join('tools', 'templates');

const { log } = console;

class Templates {
  /**
   * The template for RPC methods
   */
  method: string;
  /**
   * The template partial for RPC messages
   */
  message: string;
  /**
   * The template partial to use for request messages. This differs from
   * the 'message' template in that is contains a "REST Placement" column
   * in the table of fields
   */
  request_message: string;
  /**
   * The template for enums
   */
  enumeration: string;
  /**
   * The template for all rest endpoints
   */
  rest_endpoints: string;
  /**
   * The template for daemon summary docs
   */
  daemons_shared: string;
  /**
   * The template for the the code tabs
   */
  code_tabs: string;
  /**
   * The template for gRPC Javascript code samples
   */
  code_grpc_js: string;
  /**
   * The template for REST Javascript code samples
   */
  code_rest_js: string;
  /**
   * The template for gRPC Python code samples
   */
  code_grpc_py: string;
  /**
   * The template for REST Python code samples
   */
  code_rest_py: string;

  loadTemplate(fileName: string) {
    const filePath = path.join(TEMPLATES_DIR, `${fileName}.md`);
    log(`Loading template ${filePath}`);
    return fs.readFileSync(filePath).toString();
  }

  loadDaemonContent(name: string) {
    const filePath = path.join(TEMPLATES_DIR, 'daemons', `${name}.md`);
    if (!fs.existsSync(filePath)) return '';
    log(`Loading template ${filePath}`);
    return fs.readFileSync(filePath).toString();
  }

  load() {
    this.method = this.loadTemplate('method');
    this.message = this.loadTemplate('message');
    this.request_message = this.loadTemplate('request_message');
    this.enumeration = this.loadTemplate('enum');
    this.rest_endpoints = this.loadTemplate('rest_endpoints');
    this.daemons_shared = this.loadTemplate('daemons/shared');
    this.code_tabs = this.loadTemplate('code/tabs');
    this.code_grpc_js = this.loadTemplate('code/grpc/js');
    this.code_rest_js = this.loadTemplate('code/rest/js');
    this.code_grpc_py = this.loadTemplate('code/grpc/py');
    this.code_rest_py = this.loadTemplate('code/rest/py');

    Handlebars.registerPartial('partial_message', this.message);
    Handlebars.registerPartial('partial_request_message', this.request_message);
    Handlebars.registerPartial('partial_enum', this.enumeration);
    Handlebars.registerPartial('partial_rest_endpoints', this.rest_endpoints);
    Handlebars.registerPartial('partial_daemons_shared', this.daemons_shared);
    Handlebars.registerPartial('partial_code_tabs', this.code_tabs);
    Handlebars.registerPartial('partial_code_grpc_js', this.code_grpc_js);
    Handlebars.registerPartial('partial_code_rest_js', this.code_rest_js);
    Handlebars.registerPartial('partial_code_grpc_py', this.code_grpc_py);
    Handlebars.registerPartial('partial_code_rest_py', this.code_rest_py);

    Handlebars.registerHelper('upper', (value) => value?.toUpperCase());
    Handlebars.registerHelper('lower', (value) => value?.toLowerCase());
    Handlebars.registerHelper('camel', (value) => camelCase(value));
    Handlebars.registerHelper('snake', (value) => snakeCase(value));
    Handlebars.registerHelper('pascal', (value) => pascalCase(value));
  }
}

export const templates = new Templates();
